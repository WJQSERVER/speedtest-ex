/*
    LibreSpeed - 主程序
    作者：Federico Dossena
    https://github.com/librespeed/speedtest/
    GNU LGPLv3 许可证
*/


// 定义测速状态常量, 提高代码可读性
const SpeedtestStatus = {
    WAITING: 0,         // 等待配置或添加服务器
    ADDING_SERVERS: 1,  // 正在添加服务器
    SERVER_SELECTED: 2, // 服务器选择完成
    RUNNING: 3,         // 测试进行中
    DONE: 4,            // 测试完成
    ABORTED: 5          // 测试已中止(此状态在worker内部使用)
};

class Speedtest {
    /**
     * Speedtest构造函数
     */
    constructor() {
        this._serverList = []; // 测试节点服务器列表
        this._selectedServer = null; // 已选择的服务器
        this._settings = {}; // 传递给worker的测速设置
        this._status = SpeedtestStatus.WAITING; // 当前状态机状态
        this._worker = null; // Web Worker实例
        this._updateInterval = null; // 状态更新定时器
        this._originalExtra = undefined; // 原始的遥测附加数据

        console.log("LibreSpeed by Federico Dossena v5.4.1 - https://github.com/librespeed/speedtest");
        console.log("Refactored by WJQSERVER & Gemini-AI - https://github.com/WJQSERVER/speedtest-ex");
    }

    /**
     * 获取当前测试状态
     * @returns {number} 状态码 (参考 SpeedtestStatus)
     */
    getStatus() {
        return this._status;
    }

    /**
     * 设置测速参数
     * @param {string} parameter - 参数名
     * @param {*} value - 参数值
     */
    setParameter(parameter, value) {
        if (this._status === SpeedtestStatus.RUNNING) {
            throw new Error("Cannot change settings while test is running");
        }
        this._settings[parameter] = value;
        if (parameter === "telemetry_extra") {
            this._originalExtra = value;
        }
    }

    /**
     * 校验服务器对象定义是否合法
     * @param {object} server - 服务器对象
     * @private
     */
    _checkServerDefinition(server) {
        try {
            if (typeof server.name !== "string") throw new Error("Server definition missing 'name' string");
            if (typeof server.server !== "string") throw new Error("Server definition missing 'server' string address");
            if (server.server.slice(-1) !== "/") server.server += "/";
            if (server.server.startsWith("//")) server.server = window.location.protocol + server.server;
            if (typeof server.dlURL !== "string") throw new Error("Server definition missing 'dlURL' string");
            if (typeof server.ulURL !== "string") throw new Error("Server definition missing 'ulURL' string");
            if (typeof server.pingURL !== "string") throw new Error("Server definition missing 'pingURL' string");
            if (typeof server.getIpURL !== "string") throw new Error("Server definition missing 'getIpURL' string");
        } catch (e) {
            throw new Error(`Invalid server definition: ${e.message}`);
        }
    }

    /**
     * 添加一个测速节点
     * @param {object} server - 服务器对象
     */
    addTestPoint(server) {
        this._checkServerDefinition(server);
        if (this._status === SpeedtestStatus.WAITING) this._status = SpeedtestStatus.ADDING_SERVERS;
        if (this._status !== SpeedtestStatus.ADDING_SERVERS) throw new Error("Cannot add server after server selection");
        this._settings.mpot = true;
        this._serverList.push(server);
    }

    /**
     * 添加一个服务器对象数组作为测速节点
     * @param {Array<object>} list - 服务器对象数组
     */
    addTestPoints(list) {
        list.forEach(server => this.addTestPoint(server));
    }

    /**
     * 从指定URL加载JSON格式的服务器列表
     * @param {string} url - 服务器列表的URL
     * @param {function(Array<object>|null)} callback - 加载完成后的回调函数
     */
    loadServerList(url, callback) {
        if (this._status === SpeedtestStatus.WAITING) this._status = SpeedtestStatus.ADDING_SERVERS;
        if (this._status !== SpeedtestStatus.ADDING_SERVERS) throw new Error("Cannot add server after server selection");
        this._settings.mpot = true;

        const xhr = new XMLHttpRequest();
        xhr.withCredentials = true;
        xhr.onload = () => {
            try {
                const servers = JSON.parse(xhr.responseText);
                servers.forEach(server => this._checkServerDefinition(server));
                this.addTestPoints(servers);
                callback(servers);
            } catch (e) {
                console.error("Failed to parse server list", e);
                callback(null);
            }
        };
        xhr.onerror = () => callback(null);
        xhr.open("GET", url);
        xhr.send();
    }

    /**
     * 获取当前选定的服务器
     * @returns {object}
     */
    getSelectedServer() {
        if (this._status < SpeedtestStatus.SERVER_SELECTED || !this._selectedServer) {
            throw new Error("No server selected");
        }
        return this._selectedServer;
    }

    /**
     * 手动设置要使用的测速服务器
     * @param {object} server - 服务器对象
     */
    setSelectedServer(server) {
        this._checkServerDefinition(server);
        if (this._status === SpeedtestStatus.RUNNING) {
            throw new Error("Cannot select server while test is running");
        }
        this._selectedServer = server;
        this._status = SpeedtestStatus.SERVER_SELECTED;
    }

    /**
     * 内部方法: 对单个服务器进行ping测试
     * @param {object} server - 要ping的服务器对象
     * @param {number} pings - ping的次数
     * @returns {Promise<number>} 返回最佳ping值
     * @private
     */
    _pingServer(server) {
        return new Promise((resolve, reject) => {
            const PING_TIMEOUT = 2000;
            const PINGS_TO_PERFORM = 3; // 每个服务器ping 3次取最优值
            const SLOW_THRESHOLD = 500; // 如果ping值高于此值则停止后续ping
            let pings = [];
            let completedPings = 0;

            const performPing = () => {
                if (completedPings >= PINGS_TO_PERFORM) {
                    if (pings.length > 0) {
                        resolve(Math.min(...pings));
                    } else {
                        reject(new Error("Pings failed to return a value"));
                    }
                    return;
                }

                const url = `${server.server}${server.pingURL}${(server.pingURL.includes('?') ? '&' : '?')}cors=true`;
                const request = new XMLHttpRequest();
                request.withCredentials = true;
                let startTime = Date.now();

                request.onload = () => {
                    if (request.status === 200) {
                        let latency = Date.now() - startTime;
                        // 尝试使用更精确的 Performance API
                        try {
                            const perfEntry = performance.getEntriesByName(url).pop();
                            if (perfEntry) {
                                const perfLatency = perfEntry.responseStart - perfEntry.requestStart;
                                if (perfLatency > 0 && perfLatency < latency) {
                                    latency = perfLatency;
                                }
                            }
                        } catch (e) {
                            // 忽略错误, 使用XHR的时间
                        }
                        pings.push(latency);
                        completedPings++;
                        if (latency < SLOW_THRESHOLD) {
                            performPing(); // 如果延迟低, 继续ping
                        } else {
                            resolve(Math.min(...pings)); // 延迟高, 提前结束
                        }
                    } else {
                        reject(new Error(`Ping request failed with status ${request.status}`));
                    }
                };

                request.onerror = () => {
                    completedPings++;
                    performPing();
                };
                request.ontimeout = () => {
                    completedPings++;
                    performPing();
                };

                request.open('GET', url, true);
                request.timeout = PING_TIMEOUT;
                request.send();
            };

            performPing();
        });
    }

    /**
     * 自动选择延迟最低的服务器, 这是一个异步操作
     * @param {function(object|null)} resultCallback - 选择完成后的回调函数, 参数为选中的服务器或null
     */
    async selectServer(resultCallback) {
        if (this._status === SpeedtestStatus.WAITING) throw new Error("No test points added");
        if (this._status >= SpeedtestStatus.SERVER_SELECTED) throw new Error("Server already selected");
        if (this._status === SpeedtestStatus.RUNNING) throw new Error("Cannot select server while test is running");

        const pingPromises = this._serverList.map(server => this._pingServer(server).then(ping => ({ server, ping })).catch(() => null));

        const results = await Promise.all(pingPromises);

        const validResults = results.filter(r => r !== null);

        if (validResults.length === 0) {
            this._selectedServer = null;
            if (resultCallback) resultCallback(null);
            return;
        }

        const bestResult = validResults.reduce((best, current) => (current.ping < best.ping ? current : best));

        this._selectedServer = bestResult.server;
        this._status = SpeedtestStatus.SERVER_SELECTED;

        if (resultCallback) {
            resultCallback(this._selectedServer);
        }
    }

    /**
     * 开始测速
     */
    start() {
        if (this._status === SpeedtestStatus.RUNNING) throw new Error("Test is already running");
        if (this._status === SpeedtestStatus.WAITING || this._status === SpeedtestStatus.ADDING_SERVERS) {
            throw new Error("You must select a server before starting the test");
        }

        this._worker = new Worker("speedtest_worker.js?r=" + Math.random());

        this._worker.onmessage = (event) => {
            const data = event.data;
            if (this.onupdate) this.onupdate(data);

            if (data.testState >= SpeedtestStatus.DONE) {
                if (this._updateInterval) clearInterval(this._updateInterval);
                this._status = SpeedtestStatus.DONE;
                if (this.onend) this.onend(data.testState === SpeedtestStatus.ABORTED);
            }
        };

        this._updateInterval = setInterval(() => {
            this._worker.postMessage({ command: "status" });
        }, 200);

        if (this._status === SpeedtestStatus.SERVER_SELECTED) {
            this._settings.url_dl = this._selectedServer.server + this._selectedServer.dlURL;
            this._settings.url_ul = this._selectedServer.server + this._selectedServer.ulURL;
            this._settings.url_ping = this._selectedServer.server + this._selectedServer.pingURL;
            this._settings.url_getIp = this._selectedServer.server + this._selectedServer.getIpURL;
            
            const extra = { server: this._selectedServer.name };
            if (this._originalExtra) {
                extra.extra = this._originalExtra;
            }
            this._settings.telemetry_extra = JSON.stringify(extra);
        }

        this._status = SpeedtestStatus.RUNNING;
        this._worker.postMessage({ command: "start", settings: this._settings });
    }

    /**
     * 中止测试
     */
    abort() {
        if (this._status < SpeedtestStatus.RUNNING) {
            throw new Error("Cannot abort a test that is not running");
        }
        if (this._status === SpeedtestStatus.RUNNING && this._worker) {
            this._worker.postMessage({ command: "abort" });
        }
    }

    // 事件回调钩子, 由用户定义
    onupdate = (data) => {}; // 接收 {dlStatus, ulStatus, pingStatus, jitterStatus, ...}
    onend = (aborted) => {}; // 接收一个布尔值, 表示测试是否被中止
}