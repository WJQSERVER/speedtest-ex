function startStop() {
    if (s.getState() == 3) {
        s.abort();
        I("startStopBtn").className = "";
        I("startStopBtn").textContent = "开始";
        initUI();
    } else {
        I("startStopBtn").className = "running";
        I("startStopBtn").textContent = "停止";
        s.onupdate = function (data) {
            I("ip").textContent = data.clientIp;
            I("dlText").textContent = (data.testState == 1 && data.dlStatus == 0) ? "..." : data.dlStatus;
            I("ulText").textContent = (data.testState == 3 && data.ulStatus == 0) ? "..." : data.ulStatus;
            I("pingText").textContent = data.pingStatus;
            I("jitText").textContent = data.jitterStatus;
        };
        s.onend = function (aborted) {
            I("startStopBtn").className = "";
            I("startStopBtn").textContent = "开始";
        };
        s.start();
    }
}

function initUI() {
    I("dlText").textContent = "";
    I("ulText").textContent = "";
    I("pingText").textContent = "";
    I("jitText").textContent = "";
    I("ip").textContent = "";
    I("startStopBtn").textContent = "开始";
}

setTimeout(function () { initUI(); }, 100);

let socket;
let timeoutCount = 0;

function setupWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
    socket = new WebSocket(`${protocol}://${window.location.host}/ws`);
    socket.onopen = function() {
        console.log('WebSocket 连接已建立');
        // 可选：可以发送初始消息以开始 ping 过程
    };

    socket.onmessage = function(event) {
        const data = JSON.parse(event.data);
        const pingValueDiv = document.getElementById('pingValue');

        if (data.success) {
            pingValueDiv.textContent = data.rtt.toFixed(3); // 更新当前 Ping 值
            timeoutCount = 0; // 重置超时计数器
        } else {
            pingValueDiv.textContent = '-'; // 重置 Ping 值
            if (data.error === "timeout" || data.error === "revping-not-online") {
                timeoutCount++; // 超时计数器加一
                if (timeoutCount >= 5) { // 超时计数器达到5次，停止 WebSocket 连接
                    console.log("RevPing: 检测到 5 次超时；停止 WebSocket 连接。后端无法使用 ICMP Echo 接收来自您的 IP 的回复。/ RevPing 功能已禁用。");
                    socket.close();
                }
            } else {
                timeoutCount = 0; // 重置超时计数器
            }
        }
    };

    socket.onerror = function(error) {
        console.error('WebSocket 错误:', error);
    };

    socket.onclose = function() {
        console.log('WebSocket 连接已关闭');
    };
}

function fetchVersion() {
    fetch('/api/version')
    .then(response => response.json())
    .then(data => {
        document.getElementById('versionBadge').textContent = data.Version;
    })
    .catch(error => {
        console.error('获取版本失败:', error);
    });
}

document.addEventListener('DOMContentLoaded', fetchVersion);

// 网页加载时开始 WebSocket 连接
window.onload = setupWebSocket;