/*
    LibreSpeed - 主程序
    作者：Federico Dossena
    https://github.com/librespeed/speedtest/
    GNU LGPLv3 许可证
*/

/*
   这是您的网页和速度测试之间的主要接口。
   它对页面隐藏了速度测试的 Web Worker，并提供了许多方便的函数来控制测试。

   学习如何使用它的最佳方法是查看基本示例，但这里有一些文档。

   要初始化测试，请创建一个新的 Speedtest 对象：
    let s = new Speedtest();
   现在您可以将其视为一个有限状态机。这些是状态（使用 getState() 查看它们）：
   - 0: 在这里，您可以使用 setParameter("参数", 值) 方法更改速度测试设置（如测试持续时间）。从这里，您可以使用 start() 开始测试（转到状态 3），或者使用 addTestPoint(server) 或 addTestPoints(serverList) 添加多个测试点（转到状态 1）。此外，这是设置 onupdate(data) 和 onend(aborted) 事件回调的最佳时机。
   - 1: 在这里，您可以添加测试点。只有在要使用多个测试点时才需要这样做。
        服务器定义为如下对象：
        {
            name: "用户友好的名称",
            server: "http://yourBackend.com/",     <---- 您的服务器的 URL。您可以指定 http:// 或 https://。如果您的服务器支持两者，只需写 // 而不带协议
            dlURL: "garbage.php"    <----- 服务器上 garbage.php 或其替代品的路径
            ulURL: "empty.php"    <----- 服务器上 empty.php 或其替代品的路径
            pingURL: "empty.php"    <----- 服务器上 empty.php 或其替代品的路径。这用于通过此选择器 ping 服务器
            getIpURL: "getIP.php"    <----- 服务器上 getIP.php 或其替代品的路径
        }
        在状态 1 时，您只能添加测试点，不能更改测试设置。完成后，使用 selectServer(callback) 选择 ping 值最低的测试点。这是异步的，完成后，它将调用您的回调函数并移至状态 2。调用 setSelectedServer(server) 将手动选择服务器并移至状态 2。
    - 2: 测试点已选择，准备开始测试。使用 start() 开始，这将移至状态 3
    - 3: 测试运行中。在这里，您的 onupdate 事件回调将被定期调用，数据来自 worker，包含有关速度和进度的信息。一个数据对象将传递给您的 onupdate 函数，包含以下项目：
            - dlStatus: 下载速度（Mbit/s）
            - ulStatus: 上传速度（Mbit/s）
            - pingStatus: ping 值（ms）
            - jitterStatus: 抖动（ms）
            - dlProgress: 下载测试进度（0-1 的浮点数）
            - ulProgress: 上传测试进度（0-1 的浮点数）
            - pingProgress: ping/抖动测试进度（0-1 的浮点数）
            - testState: 测试状态（-1=未开始，0=开始中，1=下载测试，2=ping+抖动测试，3=上传测试，4=已完成，5=已中止）
            - clientIp: 执行测试的客户端 IP 地址（可选包括 ISP 和距离）
        在测试结束时，将调用 onend 函数，传递一个布尔值，指定测试是被中止还是正常结束。
        可以随时使用 abort() 中止测试。
        测试结束时，它将移至状态 4
    - 4: 测试完成。如果需要，您可以通过调用 start() 再次运行它。
 */

function Speedtest() {
  this._serverList = []; // 使用多个测试点时，这是测试点列表
  this._selectedServer = null; // 使用多个测试点时，这是选定的服务器
  this._settings = {}; // 速度测试 worker 的设置
  this._state = 0; // 0=添加设置，1=添加服务器，2=服务器选择完成，3=测试运行中，4=完成
  console.log(
    "LibreSpeed by Federico Dossena v5.4.1 - https://github.com/librespeed/speedtest"
  );
  console.log(
    "SpeedTest-EX by WJQSERVER - https://github.com/WJQSERVER/speedtest-ex"
  )
}

Speedtest.prototype = {
  constructor: Speedtest,
  /**
   * 返回测试的状态：0=添加设置，1=添加服务器，2=服务器选择完成，3=测试运行中，4=完成
   */
  getState: function () {
    return this._state;
  },
  /**
   * 更改测试设置中的一个参数，使其偏离默认值。
   * - parameter: 要设置的参数名称的字符串
   * - value: 参数的新值
   *
   * 无效的值或不存在的参数将被速度测试 worker 忽略。
   */
  setParameter: function (parameter, value) {
    if (this._state == 3)
      throw "测试运行时无法更改测试设置";
    this._settings[parameter] = value;
    if (parameter === "telemetry_extra") {
      this._originalExtra = this._settings.telemetry_extra;
    }
  },
  /**
   * 内部使用，检查服务器对象是否包含所有必需的元素。
   * 如果需要，还会修复服务器 URL。
   */
  _checkServerDefinition: function (server) {
    try {
      if (typeof server.name !== "string")
        throw "服务器定义中缺少名称字符串 (name)";
      if (typeof server.server !== "string")
        throw "服务器定义中缺少服务器地址字符串 (server)";
      if (server.server.charAt(server.server.length - 1) != "/")
        server.server += "/";
      if (server.server.indexOf("//") == 0)
        server.server = location.protocol + server.server;
      if (typeof server.dlURL !== "string")
        throw "服务器定义中缺少下载 URL 字符串 (dlURL)";
      if (typeof server.ulURL !== "string")
        throw "服务器定义中缺少上传 URL 字符串 (ulURL)";
      if (typeof server.pingURL !== "string")
        throw "服务器定义中缺少 Ping URL 字符串 (pingURL)";
      if (typeof server.getIpURL !== "string")
        throw "服务器定义中缺少 GetIP URL 字符串 (getIpURL)";
    } catch (e) {
      throw "无效的服务器定义";
    }
  },
  /**
   * 添加测试点（多个测试点）
   * server: 要添加的服务器对象。必须包含以下元素：
   *  {
   *       name: "用户友好的名称",
   *       server: "http://yourBackend.com/",   服务器的 URL。可以指定 http:// 或 https://。如果服务器支持两者，只需写 // 而不带协议
   *       dlURL: "garbage.php"   服务器上 garbage.php 或其替代品的路径
   *       ulURL: "empty.php"   服务器上 empty.php 或其替代品的路径
   *       pingURL: "empty.php"   服务器上 empty.php 或其替代品的路径。这用于通过此选择器 ping 服务器
   *       getIpURL: "getIP.php"   服务器上 getIP.php 或其替代品的路径
   *   }
   */
  addTestPoint: function (server) {
    this._checkServerDefinition(server);
    if (this._state == 0) this._state = 1;
    if (this._state != 1) throw "服务器选择后无法添加服务器";
    this._settings.mpot = true;
    this._serverList.push(server);
  },
  /**
   * 与 addTestPoint 相同，但可以传递服务器数组
   */
  addTestPoints: function (list) {
    for (let i = 0; i < list.length; i++) this.addTestPoint(list[i]);
  },
  /**
   * 从 URL 加载 JSON 服务器列表（多个测试点）
   * url: 可以获取服务器列表的 url。必须是一个数组，包含具有以下元素的对象：
   *  {
   *       "name": "用户友好的名称",
   *       "server": "http://yourBackend.com/",   服务器的 URL。可以指定 http:// 或 https://。如果服务器支持两者，只需写 // 而不带协议
   *       "dlURL": "garbage.php"   服务器上 garbage.php 或其替代品的路径
   *       "ulURL": "empty.php"   服务器上 empty.php 或其替代品的路径
   *       "pingURL": "empty.php"   服务器上 empty.php 或其替代品的路径。这用于通过此选择器 ping 服务器
   *       "getIpURL": "getIP.php"   服务器上 getIP.php 或其替代品的路径
   *   }
   * result: 列表正确加载时要调用的回调。将向此函数传递一个包含加载的服务器的数组，如果失败则为 null
   */
  loadServerList: function (url, result) {
    if (this._state == 0) this._state = 1;
    if (this._state != 1) throw "服务器选择后无法添加服务器";
    this._settings.mpot = true;
    let xhr = new XMLHttpRequest();
    xhr.onload = function () {
      try {
        const servers = JSON.parse(xhr.responseText);
        for (let i = 0; i < servers.length; i++) {
          this._checkServerDefinition(servers[i]);
        }
        this.addTestPoints(servers);
        result(servers);
      } catch (e) {
        result(null);
      }
    }.bind(this);
    xhr.onerror = function () { result(null); }
    xhr.open("GET", url);
    xhr.send();
  },
  /**
   * 返回选定的服务器（多个测试点）
   */
  getSelectedServer: function () {
    if (this._state < 2 || this._selectedServer == null)
      throw "未选择服务器";
    return this._selectedServer;
  },
  /**
   * 手动选择其中一个测试点（多个测试点）
   */
  setSelectedServer: function (server) {
    this._checkServerDefinition(server);
    if (this._state == 3)
      throw "测试运行时无法选择服务器";
    this._selectedServer = server;
    this._state = 2;
  },
  /**
   * 从添加的测试点列表中自动选择服务器。将选择 ping 值最低的服务器。（多个测试点）
   * 这个过程是异步的，传递的 result 回调函数将在完成时被调用，然后可以开始测试。
   */
  selectServer: function (result) {
    if (this._state != 1) {
      if (this._state == 0) throw "未添加测试点";
      if (this._state == 2) throw "服务器已选择";
      if (this._state >= 3)
        throw "测试运行时无法选择服务器";
    }
    if (this._selectServerCalled) throw "selectServer 已被调用"; else this._selectServerCalled = true;
    /*这个函数遍历服务器列表。对于每个服务器，测量 ping 值，然后用最佳服务器调用选定的函数，如果所有服务器都不可用，则为 null。
     */
    const select = function (serverList, selected) {
      //ping 指定的 URL，然后调用 result 函数。result 将接收一个参数，该参数要么是 ping URL 所需的时间，要么在出错时为 -1。
      const PING_TIMEOUT = 2000;
      let USE_PING_TIMEOUT = true; //将在不支持的浏览器上禁用
      if (/MSIE.(\d+\.\d+)/i.test(navigator.userAgent)) {
        //IE11 不支持 XHR 超时
        USE_PING_TIMEOUT = false;
      }
      const ping = function (url, rtt) {
        url += (url.match(/\?/) ? "&" : "?") + "cors=true";
        let xhr = new XMLHttpRequest();
        let t = new Date().getTime();
        xhr.onload = function () {
          if (xhr.responseText.length == 0) {
            //我们期望一个空响应
            let instspd = new Date().getTime() - t; //粗略的时间估计
            try {
              //尝试使用性能 API 获取更准确的时间
              let p = performance.getEntriesByName(url);
              p = p[p.length - 1];
              let d = p.responseStart - p.requestStart;
              if (d <= 0) d = p.duration;
              if (d > 0 && d < instspd) instspd = d;
            } catch (e) { }
            rtt(instspd);
          } else rtt(-1);
        }.bind(this);
        xhr.onerror = function () {
          rtt(-1);
        }.bind(this);
        xhr.open("GET", url);
        if (USE_PING_TIMEOUT) {
          try {
            xhr.timeout = PING_TIMEOUT;
            xhr.ontimeout = xhr.onerror;
          } catch (e) { }
        }
        xhr.send();
      }.bind(this);

      //这个函数重复 ping 服务器以获得良好的 ping 估计。完成后，它调用 done 函数，不带参数。执行结束时，服务器将有一个新参数 pingT，它要么是我们从服务器获得的最佳 ping，要么在出错时为 -1。
      const PINGS = 3, //最多执行 3 次 ping，除非服务器宕机...
        SLOW_THRESHOLD = 500; //...或者其中一个 ping 高于此阈值
      const checkServer = function (server, done) {
        let i = 0;
        server.pingT = -1;
        if (server.server.indexOf(location.protocol) == -1) done();
        else {
          const nextPing = function () {
            if (i++ == PINGS) {
              done();
              return;
            }
            ping(
              server.server + server.pingURL,
              function (t) {
                if (t >= 0) {
                  if (t < server.pingT || server.pingT == -1) server.pingT = t;
                  if (t < SLOW_THRESHOLD) nextPing();
                  else done();
                } else done();
              }.bind(this)
            );
          }.bind(this);
          nextPing();
        }
      }.bind(this);
      //逐个检查列表中的服务器
      let i = 0;
      const done = function () {
        let bestServer = null;
        for (let i = 0; i < serverList.length; i++) {
          if (
            serverList[i].pingT != -1 &&
            (bestServer == null || serverList[i].pingT < bestServer.pingT)
          )
            bestServer = serverList[i];
        }
        selected(bestServer);
      }.bind(this);
      const nextServer = function () {
        if (i == serverList.length) {
          done();
          return;
        }
        checkServer(serverList[i++], nextServer);
      }.bind(this);
      nextServer();
    }.bind(this);

    //并行服务器选择
    const CONCURRENCY = 6;
    let serverLists = [];
    for (let i = 0; i < CONCURRENCY; i++) {
      serverLists[i] = [];
    }
    for (let i = 0; i < this._serverList.length; i++) {
      serverLists[i % CONCURRENCY].push(this._serverList[i]);
    }
    let completed = 0;
    let bestServer = null;
    for (let i = 0; i < CONCURRENCY; i++) {
      select(
        serverLists[i],
        function (server) {
          if (server != null) {
            if (bestServer == null || server.pingT < bestServer.pingT)
              bestServer = server;
          }
          completed++;
          if (completed == CONCURRENCY) {
            this._selectedServer = bestServer;
            this._state = 2;
            if (result) result(bestServer);
          }
        }.bind(this)
      );
    }
  },
  /**
   * 开始测试。
   * 在测试期间，onupdate(data) 回调函数将定期被调用，传递来自 worker 的数据。
   * 在测试结束时，将调用 onend(aborted) 函数，传递一个布尔值，告诉您测试是被中止还是正常结束。
   */
  start: function () {
    if (this._state == 3) throw "测试已在运行";
    this.worker = new Worker("speedtest_worker.js?r=" + Math.random());
    this.worker.onmessage = function (e) {
      if (e.data === this._prevData) return;
      else this._prevData = e.data;
      const data = JSON.parse(e.data);
      try {
        if (this.onupdate) this.onupdate(data);
      } catch (e) {
        console.error("Speedtest onupdate 事件抛出异常: " + e);
      }
      if (data.testState >= 4) {
        clearInterval(this.updater);
        this._state = 4;
        try {
          if (this.onend) this.onend(data.testState == 5);
        } catch (e) {
          console.error("Speedtest onend 事件抛出异常: " + e);
        }
      }
    }.bind(this);
    this.updater = setInterval(
      function () {
        this.worker.postMessage("status");
      }.bind(this),
      200
    );
    if (this._state == 1)
      throw "使用多个测试点时，必须在开始测试前调用 selectServer";
    if (this._state == 2) {
      this._settings.url_dl =
        this._selectedServer.server + this._selectedServer.dlURL;
      this._settings.url_ul =
        this._selectedServer.server + this._selectedServer.ulURL;
      this._settings.url_ping =
        this._selectedServer.server + this._selectedServer.pingURL;
      this._settings.url_getIp =
        this._selectedServer.server + this._selectedServer.getIpURL;
      if (typeof this._originalExtra !== "undefined") {
        this._settings.telemetry_extra = JSON.stringify({
          server: this._selectedServer.name,
          extra: this._originalExtra
        });
      } else
        this._settings.telemetry_extra = JSON.stringify({
          server: this._selectedServer.name
        });
    }
    this._state = 3;
    this.worker.postMessage("start " + JSON.stringify(this._settings));
  },
  /**
   * 在测试运行时中止测试。
   */
  abort: function () {
    if (this._state < 3) throw "无法中止尚未开始的测试";
    if (this._state < 4) this.worker.postMessage("abort");
  }
};