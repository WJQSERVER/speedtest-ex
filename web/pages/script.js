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

let intervalId;

async function icmpping() {
    const response = await fetch(`/revping`);
    const data = await response.json();
    const resultDiv = document.getElementById('result');
    const pingValueDiv = document.getElementById('pingValue');

    if (data.success) {
        /* icmpresultDiv.innerHTML = `
             <p> ${data.rtt.toFixed(3)} ms</p>
         `; */
        pingValueDiv.textContent = data.rtt.toFixed(3); // 更新当前 Ping 值
    } else {
        /* icmpresultDiv.innerHTML = `<p style="color: red;">错误: ${data.error}</p>`; */
        pingValueDiv.textContent = '-'; // 重置 Ping 值
    }
}

function startIcmpPing() {
    icmpping(); // 立即请求一次
    intervalId = setInterval(icmpping, 2500); // 每隔2.5秒请求一次
}

// 网页加载时开始请求
window.onload = startIcmpPing;