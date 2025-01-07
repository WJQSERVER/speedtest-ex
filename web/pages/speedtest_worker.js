/*
	LibreSpeed - Worker
	作者：Federico Dossena
	https://github.com/librespeed/speedtest/
	GNU LGPLv3 许可证
*/

// 报告给主线程的数据
let testState = -1; // -1=未开始, 0=正在开始, 1=下载测试, 2=ping+抖动测试, 3=上传测试, 4=已完成, 5=中止
let dlStatus = ""; // 下载速度，单位为兆比特/秒，保留两位小数
let ulStatus = ""; // 上传速度，单位为兆比特/秒，保留两位小数
let pingStatus = ""; // ping值，单位为毫秒，保留两位小数
let jitterStatus = ""; // 抖动值，单位为毫秒，保留两位小数
let clientIp = ""; // 客户端IP地址，由getIP.php报告
let dlProgress = 0; // 下载测试进度 0-1
let ulProgress = 0; // 上传测试进度 0-1
let pingProgress = 0; // ping+抖动测试进度 0-1
let testId = null; // 测试ID（如果使用遥测，则由遥测返回，否则为null）

let log = ""; // 遥测日志
function tlog(s) {
	if (settings.telemetry_level >= 2) {
		log += Date.now() + ": " + s + "\n";
	}
}
function tverb(s) {
	if (settings.telemetry_level >= 3) {
		log += Date.now() + ": " + s + "\n";
	}
}
function twarn(s) {
	if (settings.telemetry_level >= 2) {
		log += Date.now() + " 警告: " + s + "\n";
	}
	console.warn(s);
}

// 测试设置。可以通过在开始命令中发送特定值来覆盖
let settings = {
	mpot: false, // 设置为true时处于MPOT模式
	test_order: "IP_D_U", // 执行测试的顺序，以字符串形式。D=下载, U=上传, P=Ping+抖动, I=IP, _=1秒延迟
	time_ul_max: 15, // 上传测试的最大持续时间（秒）
	time_dl_max: 15, // 下载测试的最大持续时间（秒）
	time_auto: true, // 如果设置为true，在更快的连接上测试会花费更少的时间
	time_ulGraceTime: 3, // 在实际测量上传速度之前等待的时间（秒）（等待缓冲区填满）
	time_dlGraceTime: 1.5, // 在实际测量下载速度之前等待的时间（秒）（等待TCP窗口增加）
	count_ping: 10, // 在ping测试中执行的ping次数
	url_dl: "backend/garbage", // 用于下载测试的大文件或garbage.php的路径
	url_ul: "backend/empty", // 用于上传测试的空文件的路径
	url_ping: "backend/empty", // 用于ping测试的空文件的路径
	url_getIp: "backend/getIP", // getIP.php的路径或类似的输出客户端IP的文件
	getIp_ispInfo: true, // 如果设置为true，服务器将包含ISP信息和IP地址
	getIp_ispInfo_distance: "km", // km或mi=估计与服务器的距离（公里/英里）；设置为false以禁用距离估计。必须启用getIp_ispInfo才能使用此功能
	xhr_dlMultistream: 6, // 要使用的下载流数量（如果enable_quirks处于活动状态，可能会不同）
	xhr_ulMultistream: 3, // 要使用的上传流数量（如果enable_quirks处于活动状态，可能会不同）
	xhr_multistreamDelay: 300, // 并发请求应该延迟多少
	xhr_ignoreErrors: 1, // 0=失败时中止, 1=尝试重新启动失败的流, 2=忽略所有错误
	xhr_dlUseBlob: false, // 如果设置为true，它会减少RAM使用量，但会使用硬盘（在大型garbagePhp_chunkSize和/或高xhr_dlMultistream时有用）
	xhr_ul_blob_megabytes: 20, // 上传测试中发送的上传blob的大小（MB）（在chrome移动版上强制为4）
	garbagePhp_chunkSize: 100, // garbage.php发送的块大小（如果enable_quirks处于活动状态，可能会不同）
	enable_quirks: true, // 为特定浏览器启用quirks。目前它会覆盖设置以优化特定浏览器，除非它们已经被start命令覆盖
	ping_allowPerformanceApi: true, // 如果启用，ping测试将尝试使用Performance API更精确地计算ping。目前在Chrome中完美工作，在Edge中效果不佳，在Firefox中完全不工作。如果不支持Performance API或结果明显错误，将提供一个后备方案。
	overheadCompensationFactor: 1.06, // 可以更改以补偿传输开销。（参见doc.md了解其他一些值）
	useMebibits: false, // 如果设置为true，速度将以mebibits/s而不是megabits/s为单位报告
	telemetry_level: 0, // 0=禁用, 1=基本 (仅结果), 2=完整 (结果和时间) 3=调试 (结果+日志)
	url_telemetry: "results/telemetry", // 将遥测数据添加到数据库的脚本的路径
	telemetry_extra: "", // 可以通过设置传递给遥测的额外数据
	forceIE11Workaround: false // 当设置为true时，它将在所有浏览器上强制使用IE11上传测试。仅用于调试
};

let xhr = null; // 当前活动的xhr请求数组
let interval = null; // 测试中使用的定时器
let test_pointer = 0; // 指向settings.test_order中下一个要运行的测试的指针

/*
  这个函数用于在设置中传递的URL上确定我们是否需要使用?还是&作为分隔符
*/
function url_sep(url) {
	return url.match(/\?/) ? "&" : "?";
}

/*
	来自主线程到此worker的命令的监听器。
	命令：
	-status: 以JSON字符串的形式返回当前状态，包含testState, dlStatus, ulStatus, pingStatus, clientIp, jitterStatus, dlProgress, ulProgress, pingProgress
	-abort: 中止当前测试
	-start: 开始测试。可以选择以JSON形式传递设置。
		例如：start {"time_ul_max":"10", "time_dl_max":"10", "count_ping":"50"}
*/
this.addEventListener("message", function (e) {
	const params = e.data.split(" ");
	if (params[0] === "status") {
		// 返回状态
		postMessage(
			JSON.stringify({
				testState: testState,
				dlStatus: dlStatus,
				ulStatus: ulStatus,
				pingStatus: pingStatus,
				clientIp: clientIp,
				jitterStatus: jitterStatus,
				dlProgress: dlProgress,
				ulProgress: ulProgress,
				pingProgress: pingProgress,
				testId: testId
			})
		);
	}
	if (params[0] === "start" && testState === -1) {
		// 开始新测试
		testState = 0;
		try {
			// 解析设置（如果存在）
			let s = {};
			try {
				const ss = e.data.substring(5);
				if (ss) s = JSON.parse(ss);
			} catch (e) {
				twarn("解析自定义设置JSON时出错。请检查您的语法");
			}
			// 复制自定义设置
			for (let key in s) {
				if (typeof settings[key] !== "undefined") settings[key] = s[key];
				else twarn("忽略未知设置: " + key);
			}
			const ua = navigator.userAgent;
			// 特定浏览器的quirks。仅在未被覆盖时应用。未来的版本可能会添加更多
			if (settings.enable_quirks || (typeof s.enable_quirks !== "undefined" && s.enable_quirks)) {
				if (/Firefox.(\d+\.\d+)/i.test(ua)) {
					if (typeof s.ping_allowPerformanceApi === "undefined") {
						// ff性能API不佳
						settings.ping_allowPerformanceApi = false;
					}
				}
				if (/Edge.(\d+\.\d+)/i.test(ua)) {
					if (typeof s.xhr_dlMultistream === "undefined") {
						// edge使用3个下载流更精确
						settings.xhr_dlMultistream = 3;
					}
				}
				if (/Chrome.(\d+)/i.test(ua) && !!self.fetch) {
					if (typeof s.xhr_dlMultistream === "undefined") {
						// chrome使用5个流更精确
						settings.xhr_dlMultistream = 5;
					}
				}
			}
			if (/Edge.(\d+\.\d+)/i.test(ua)) {
				//Edge 15引入了一个bug，导致onprogress事件不被触发，我们必须使用"小块"解决方案，这会降低精度
				settings.forceIE11Workaround = true;
			}
			if (/PlayStation 4.(\d+\.\d+)/i.test(ua)) {
				//PS4浏览器有与IE11/Edge相同的bug
				settings.forceIE11Workaround = true;
			}
			if (/Chrome.(\d+)/i.test(ua) && /Android|iPhone|iPad|iPod|Windows Phone/i.test(ua)) {
				//便宜的解决方案
				//Chrome移动版在65版本左右引入了一个限制，我们必须将XHR上传大小限制为4兆字节
				settings.xhr_ul_blob_megabytes = 4;
			}
			if (/^((?!chrome|android|crios|fxios).)*safari/i.test(ua)) {
				//Safari也需要IE11解决方案，但仅适用于MPOT版本
				settings.forceIE11Workaround = true;
			}
			//telemetry_level需要解析而不是直接复制
			if (typeof s.telemetry_level !== "undefined") settings.telemetry_level = s.telemetry_level === "basic" ? 1 : s.telemetry_level === "full" ? 2 : s.telemetry_level === "debug" ? 3 : 0; // 遥测级别
			//将test_order转换为大写，以防万一
			settings.test_order = settings.test_order.toUpperCase();
		} catch (e) {
			twarn("自定义测试设置中可能存在错误。某些设置可能未被应用。异常: " + e);
		}
		// 运行测试
		tverb(JSON.stringify(settings));
		test_pointer = 0;
		let iRun = false,
			dRun = false,
			uRun = false,
			pRun = false;
		const runNextTest = function () {
			if (testState == 5) return;
			if (test_pointer >= settings.test_order.length) {
				//测试完成
				if (settings.telemetry_level > 0)
					sendTelemetry(function (id) {
						testState = 4;
						if (id != null) testId = id;
					});
				else testState = 4;
				return;
			}
			switch (settings.test_order.charAt(test_pointer)) {
				case "I":
					{
						test_pointer++;
						if (iRun) {
							runNextTest();
							return;
						} else iRun = true;
						getIp(runNextTest);
					}
					break;
				case "D":
					{
						test_pointer++;
						if (dRun) {
							runNextTest();
							return;
						} else dRun = true;
						testState = 1;
						dlTest(runNextTest);
					}
					break;
				case "U":
					{
						test_pointer++;
						if (uRun) {
							runNextTest();
							return;
						} else uRun = true;
						testState = 3;
						ulTest(runNextTest);
					}
					break;
				case "P":
					{
						test_pointer++;
						if (pRun) {
							runNextTest();
							return;
						} else pRun = true;
						testState = 2;
						pingTest(runNextTest);
					}
					break;
				case "_":
					{
						test_pointer++;
						setTimeout(runNextTest, 1000);
					}
					break;
				default:
					test_pointer++;
			}
		};
		runNextTest();
	}
	if (params[0] === "abort") {
		// 中止命令
		if (testState >= 4) return;
		tlog("手动中止");
		clearRequests(); // 停止所有xhr活动
		runNextTest = null;
		if (interval) clearInterval(interval); // 如果存在，清除定时器
		if (settings.telemetry_level > 1) sendTelemetry(function () { });
		testState = 5; // 将测试设置为中止状态
		dlStatus = "";
		ulStatus = "";
		pingStatus = "";
		jitterStatus = "";
		clientIp = "";
		dlProgress = 0;
		ulProgress = 0;
		pingProgress = 0;
	}
});

// 积极停止所有XHR活动
function clearRequests() {
	tverb("停止待处理的XHR");
	if (xhr) {
		for (let i = 0; i < xhr.length; i++) {
			try {
				xhr[i].onprogress = null;
				xhr[i].onload = null;
				xhr[i].onerror = null;
			} catch (e) { }
			try {
				xhr[i].upload.onprogress = null;
				xhr[i].upload.onload = null;
				xhr[i].upload.onerror = null;
			} catch (e) { }
			try {
				xhr[i].abort();
			} catch (e) { }
			try {
				delete xhr[i];
			} catch (e) { }
		}
		xhr = null;
	}
}

// 使用url_getIp获取客户端的IP，然后调用done函数
let ipCalled = false; // 用于防止多次意外调用getIp
let ispInfo = ""; // 用于遥测
function getIp(done) {
	tverb("getIp");
	if (ipCalled) return;
	else ipCalled = true; // getIp是否已被调用？
	let startT = new Date().getTime();
	xhr = new XMLHttpRequest();
	xhr.withCredentials = true;
	xhr.onload = function () {
		tlog("IP: " + xhr.responseText + "，耗时 " + (new Date().getTime() - startT) + "ms");
		try {
			const data = JSON.parse(xhr.responseText);
			clientIp = data.processedString;
			ispInfo = data.rawIspInfo;
		} catch (e) {
			clientIp = xhr.responseText;
			ispInfo = "";
		}
		done();
	};
	xhr.onerror = function () {
		tlog("getIp失败，耗时 " + (new Date().getTime() - startT) + "ms");
		done();
	};
	xhr.open("GET", settings.url_getIp + url_sep(settings.url_getIp) + (settings.mpot ? "cors=true&" : "") + (settings.getIp_ispInfo ? "isp=true" + (settings.getIp_ispInfo_distance ? "&distance=" + settings.getIp_ispInfo_distance + "&" : "&") : "&") + "r=" + Math.random(), true);
	xhr.send();
}

// 下载测试，完成时调用done函数
let dlCalled = false; // 用于防止多次意外调用dlTest
function dlTest(done) {
	tverb("dlTest");
	if (dlCalled) return;
	else dlCalled = true; // dlTest是否已被调用？
	let totLoaded = 0.0, // 已加载字节的总数
		startT = new Date().getTime(), // 测试开始的时间戳
		bonusT = 0, // 测试缩短的毫秒数（在更快的连接上更高）
		graceTimeDone = false, // 宽限时间过后设置为true
		failed = false; // 如果流失败则设置为true
	xhr = [];
	// 创建下载流的函数。流略微延迟，以便它们不会同时结束
	const testStream = function (i, delay) {
		setTimeout(
			function () {
				if (testState !== 1) return; // 延迟的流在下载测试结束后开始
				tverb("dl测试流开始 " + i + " " + delay);
				let prevLoaded = 0; // 上次调用onprogress时加载的字节数
				let x = new XMLHttpRequest();
				x.withCredentials = true;
				xhr[i] = x;
				xhr[i].onprogress = function (event) {
					tverb("dl流进度事件 " + i + " " + event.loaded);
					if (testState !== 1) {
						try {
							x.abort();
						} catch (e) { }
					} // 以防这个XHR在下载测试后仍在运行
					// 进度事件，将新加载的字节数添加到totLoaded
					const loadDiff = event.loaded <= 0 ? 0 : event.loaded - prevLoaded;
					if (isNaN(loadDiff) || !isFinite(loadDiff) || loadDiff < 0) return; // 以防万一
					totLoaded += loadDiff;
					prevLoaded = event.loaded;
				}.bind(this);
				xhr[i].onload = function () {
					// 大文件已完全加载，重新开始
					tverb("dl流完成 " + i);
					try {
						xhr[i].abort();
					} catch (e) { } // 重置流数据到空RAM
					testStream(i, 0);
				}.bind(this);
				xhr[i].onerror = function () {
					// 错误
					tverb("dl流失败 " + i);
					if (settings.xhr_ignoreErrors === 0) failed = true; // 中止
					try {
						xhr[i].abort();
					} catch (e) { }
					delete xhr[i];
					if (settings.xhr_ignoreErrors === 1) testStream(i, 0); // 重启流
				}.bind(this);
				// 发送xhr
				try {
					if (settings.xhr_dlUseBlob) xhr[i].responseType = "blob";
					else xhr[i].responseType = "arraybuffer";
				} catch (e) { }
				xhr[i].open("GET", settings.url_dl + url_sep(settings.url_dl) + (settings.mpot ? "cors=true&" : "") + "r=" + Math.random() + "&ckSize=" + settings.garbagePhp_chunkSize, true); // 随机字符串防止缓存
				xhr[i].send();
			}.bind(this),
			1 + delay
		);
	}.bind(this);
	// 开启流
	console.log("开始下载测试，共" + settings.xhr_dlMultistream + "个流");
	for (let i = 0; i < settings.xhr_dlMultistream; i++) {
		testStream(i, settings.xhr_multistreamDelay * i);
	}
	// 每200ms更新dlStatus
	interval = setInterval(
		function () {
			tverb("DL: " + dlStatus + (graceTimeDone ? "" : " (在宽限时间内)"));
			const t = new Date().getTime() - startT;
			if (graceTimeDone) dlProgress = (t + bonusT) / (settings.time_dl_max * 1000);
			if (t < 200) return;
			if (!graceTimeDone) {
				if (t > 1000 * settings.time_dlGraceTime) {
					if (totLoaded > 0) {
						// 如果连接太慢以至于我们还没有得到一个块，就不要重置
						startT = new Date().getTime();
						bonusT = 0;
						totLoaded = 0.0;
					}
					graceTimeDone = true;
				}
			} else {
				const speed = totLoaded / (t / 1000.0);
				if (settings.time_auto) {
					// 决定缩短测试多少。每200ms，测试缩短这里计算的bonusT
					const bonus = (5.0 * speed) / 100000;
					bonusT += bonus > 400 ? 400 : bonus;
				}
				// 更新状态
				dlStatus = ((speed * 8 * settings.overheadCompensationFactor) / (settings.useMebibits ? 1048576 : 1000000)).toFixed(2); // 速度乘以8从字节转换为比特，应用开销补偿，然后除以1048576或1000000转换为兆比特/秒或兆字节/秒
				if ((t + bonusT) / 1000.0 > settings.time_dl_max || failed) {
					// 测试结束，停止流和定时器
					if (failed || isNaN(dlStatus)) dlStatus = "失败";
					clearRequests();
					clearInterval(interval);
					dlProgress = 1;
					tlog("dlTest: " + dlStatus + "，耗时 " + (new Date().getTime() - startT) + "ms");
					done();
				}
			}
		}.bind(this),
		200
	);
}

// 上传测试，完成时调用done函数
let ulCalled = false; // 用于防止多次意外调用ulTest
function ulTest(done) {
	tverb("ulTest");
	if (ulCalled) return;
	else ulCalled = true; // ulTest是否已被调用？
	// 上传测试的垃圾数据
	let r = new ArrayBuffer(1048576);
	const maxInt = Math.pow(2, 32) - 1;
	try {
		r = new Uint32Array(r);
		for (let i = 0; i < r.length; i++) r[i] = Math.random() * maxInt;
	} catch (e) { }
	let req = [];
	let reqsmall = [];
	for (let i = 0; i < settings.xhr_ul_blob_megabytes; i++) req.push(r);
	req = new Blob(req);
	r = new ArrayBuffer(262144);
	try {
		r = new Uint32Array(r);
		for (let i = 0; i < r.length; i++) r[i] = Math.random() * maxInt;
	} catch (e) { }
	reqsmall.push(r);
	reqsmall = new Blob(reqsmall);
	const testFunction = function () {
		let totLoaded = 0.0, // 传输字节的总数
			startT = new Date().getTime(), // 测试开始的时间戳
			bonusT = 0, // 测试缩短的毫秒数（在更快的连接上更高）
			graceTimeDone = false, // 宽限时间过后设置为true
			failed = false; // 如果流失败则设置为true
		xhr = [];
		// 创建上传流的函数。流略微延迟，以便它们不会同时结束
		const testStream = function (i, delay) {
			setTimeout(
				function () {
					if (testState !== 3) return; // 延迟的流在上传测试结束后开始
					tverb("ul测试流开始 " + i + " " + delay);
					let prevLoaded = 0; // 上次调用onprogress时传输的字节数
					let x = new XMLHttpRequest();
					x.withCredentials = true;
					xhr[i] = x;
					let ie11workaround;
					if (settings.forceIE11Workaround) ie11workaround = true;
					else {
						try {
							xhr[i].upload.onprogress;
							ie11workaround = false;
						} catch (e) {
							ie11workaround = true;
						}
					}
					if (ie11workaround) {
						// IE11解决方案：xhr.upload不能正常工作，因此我们发送一堆小的256k请求并使用onload事件作为进度。这不精确，特别是在快速连接上
						xhr[i].onload = xhr[i].onerror = function () {
							tverb("ul流进度事件 (ie11wa)");
							totLoaded += reqsmall.size;
							testStream(i, 0);
						};
						xhr[i].open("POST", settings.url_ul + url_sep(settings.url_ul) + (settings.mpot ? "cors=true&" : "") + "r=" + Math.random(), true); // 随机字符串防止缓存
						try {
							xhr[i].setRequestHeader("Content-Encoding", "identity"); // 禁用压缩（某些浏览器可能会拒绝，但数据无论如何都是不可压缩的）
						} catch (e) { }
						// MPOT分支中没有Content-Type头，因为它在某些浏览器中触发bug
						xhr[i].send(reqsmall);
					} else {
						// 常规版本，无需解决方案
						xhr[i].upload.onprogress = function (event) {
							tverb("ul流进度事件 " + i + " " + event.loaded);
							if (testState !== 3) {
								try {
									x.abort();
								} catch (e) { }
							} // 以防这个XHR在上传测试后仍在运行
							// 进度事件，将新传输的字节数添加到totLoaded
							const loadDiff = event.loaded <= 0 ? 0 : event.loaded - prevLoaded;
							if (isNaN(loadDiff) || !isFinite(loadDiff) || loadDiff < 0) return; // 以防万一
							totLoaded += loadDiff;
							prevLoaded = event.loaded;
						}.bind(this);
						xhr[i].upload.onload = function () {
							// 这个流发送了所有垃圾数据，重新开始
							tverb("ul流完成 " + i);
							testStream(i, 0);
						}.bind(this);
						xhr[i].upload.onerror = function () {
							tverb("ul流失败 " + i);
							if (settings.xhr_ignoreErrors === 0) failed = true; // 中止
							try {
								xhr[i].abort();
							} catch (e) { }
							delete xhr[i];
							if (settings.xhr_ignoreErrors === 1) testStream(i, 0); // 重启流
						}.bind(this);
						// 发送xhr
						xhr[i].open("POST", settings.url_ul + url_sep(settings.url_ul) + (settings.mpot ? "cors=true&" : "") + "r=" + Math.random(), true); // 随机字符串防止缓存
						try {
							xhr[i].setRequestHeader("Content-Encoding", "identity"); // 禁用压缩（某些浏览器可能会拒绝，但数据无论如何都是不可压缩的）
						} catch (e) { }
						// MPOT分支中没有Content-Type头，因为它在某些浏览器中触发bug
						xhr[i].send(req);
					}
				}.bind(this),
				delay
			);
		}.bind(this);
		// 开启流
		for (let i = 0; i < settings.xhr_ulMultistream; i++) {
			testStream(i, settings.xhr_multistreamDelay * i);
		}
		// 每200ms更新ulStatus
		interval = setInterval(
			function () {
				tverb("UL: " + ulStatus + (graceTimeDone ? "" : " (在宽限时间内)"));
				const t = new Date().getTime() - startT;
				if (graceTimeDone) ulProgress = (t + bonusT) / (settings.time_ul_max * 1000);
				if (t < 200) return;
				if (!graceTimeDone) {
					if (t > 1000 * settings.time_ulGraceTime) {
						if (totLoaded > 0) {
							// 如果连接太慢以至于我们还没有发送一个块，就不要重置
							startT = new Date().getTime();
							bonusT = 0;
							totLoaded = 0.0;
						}
						graceTimeDone = true;
					}
				} else {
					const speed = totLoaded / (t / 1000.0);
					if (settings.time_auto) {
						// 决定缩短测试多少。每200ms，测试缩短这里计算的bonusT
						const bonus = (5.0 * speed) / 100000;
						bonusT += bonus > 400 ? 400 : bonus;
					}
					// 更新状态
					ulStatus = ((speed * 8 * settings.overheadCompensationFactor) / (settings.useMebibits ? 1048576 : 1000000)).toFixed(2); // 速度乘以8从字节转换为比特，应用开销补偿，然后除以1048576或1000000转换为兆比特/秒或兆字节/秒
					if ((t + bonusT) / 1000.0 > settings.time_ul_max || failed) {
						// 测试结束，停止流和定时器
						if (failed || isNaN(ulStatus)) ulStatus = "失败";
						clearRequests();
						clearInterval(interval);
						ulProgress = 1;
						tlog("ulTest: " + ulStatus + "，耗时 " + (new Date().getTime() - startT) + "ms");
						done();
					}
				}
			}.bind(this),
			200
		);
	}.bind(this);
	if (settings.mpot) {
		tverb("在执行上传测试之前发送POST请求");
		xhr = [];
		xhr[0] = new XMLHttpRequest();
		xhr[0].withCredentials = true;
		xhr[0].onload = xhr[0].onerror = function () {
			tverb("POST请求已发送，开始上传测试");
			testFunction();
		}.bind(this);
		xhr[0].open("POST", settings.url_ul);
		xhr[0].send();
	} else testFunction();
}

// ping+抖动测试，完成时调用done函数
let ptCalled = false; // 用于防止多次意外调用pingTest
function pingTest(done) {
	tverb("pingTest");
	if (ptCalled) return;
	else ptCalled = true; // pingTest是否已被调用？
	const startT = new Date().getTime(); // 测试开始的时间
	let prevT = null; // 上次收到pong的时间
	let ping = 0.0; // 当前ping值
	let jitter = 0.0; // 当前抖动值
	let i = 0; // 收到的pong计数器
	let prevInstspd = 0; // 上次ping时间，用于计算抖动
	xhr = [];
	// ping函数
	const doPing = function () {
		tverb("ping");
		pingProgress = i / settings.count_ping;
		prevT = new Date().getTime();
		xhr[0] = new XMLHttpRequest();
		xhr[0].withCredentials = true;
		xhr[0].onload = function () {
			// pong
			tverb("pong");
			if (i === 0) {
				prevT = new Date().getTime(); // 第一个pong
			} else {
				let instspd = new Date().getTime() - prevT;
				if (settings.ping_allowPerformanceApi) {
					try {
						// 尝试使用performance api获取准确的性能计时
						let p = performance.getEntries();
						p = p[p.length - 1];
						let d = p.responseStart - p.requestStart;
						if (d <= 0) d = p.duration;
						if (d > 0 && d < instspd) instspd = d;
					} catch (e) {
						// 如果不可能，保持估计值
						tverb("不支持Performance API，使用估计值");
					}
				}
				// 注意到一些浏览器随机有0ms ping
				if (instspd < 1) instspd = prevInstspd;
				if (instspd < 1) instspd = 1;
				const instjitter = Math.abs(instspd - prevInstspd);
				if (i === 1) ping = instspd;
				/* 第一个ping，还不能判断抖动 */ else {
					if (instspd < ping) ping = instspd; // 如果瞬时ping更低，更新ping
					if (i === 2) jitter = instjitter;
					// 丢弃第一个抖动测量，因为它可能比应有的高得多
					else jitter = instjitter > jitter ? jitter * 0.3 + instjitter * 0.7 : jitter * 0.8 + instjitter * 0.2; // 更新抖动，加权平均。ping值的峰值被赋予更多权重。
				}
				prevInstspd = instspd;
			}
			pingStatus = ping.toFixed(2);
			jitterStatus = jitter.toFixed(2);
			i++;
			tverb("ping: " + pingStatus + " 抖动: " + jitterStatus);
			if (i < settings.count_ping) doPing();
			else {
				// 还有更多ping要做吗？
				pingProgress = 1;
				tlog("ping: " + pingStatus + " 抖动: " + jitterStatus + "，耗时 " + (new Date().getTime() - startT) + "ms");
				done();
			}
		}.bind(this);
		xhr[0].onerror = function () {
			// ping失败，取消测试
			tverb("ping失败");
			if (settings.xhr_ignoreErrors === 0) {
				// 中止
				pingStatus = "失败";
				jitterStatus = "失败";
				clearRequests();
				tlog("ping测试失败，耗时 " + (new Date().getTime() - startT) + "ms");
				pingProgress = 1;
				done();
			}
			if (settings.xhr_ignoreErrors === 1) doPing(); // 重试ping
			if (settings.xhr_ignoreErrors === 2) {
				// 忽略失败的ping
				i++;
				if (i < settings.count_ping) doPing();
				else {
					// 还有更多ping要做吗？
					pingProgress = 1;
					tlog("ping: " + pingStatus + " 抖动: " + jitterStatus + "，耗时 " + (new Date().getTime() - startT) + "ms");
					done();
				}
			}
		}.bind(this);
		// 发送xhr
		xhr[0].open("GET", settings.url_ping + url_sep(settings.url_ping) + (settings.mpot ? "cors=true&" : "") + "r=" + Math.random(), true); // 随机字符串防止缓存
		xhr[0].send();
	}.bind(this);
	doPing(); // 开始第一个ping
}

// 遥测
function sendTelemetry(done) {
	if (settings.telemetry_level < 1) return;
	xhr = new XMLHttpRequest();
	xhr.withCredentials = true;
	xhr.onload = function () {
		try {
			const parts = xhr.responseText.split(" ");
			if (parts[0] == "id") {
				try {
					let id = parts[1];
					done(id);
				} catch (e) {
					done(null);
				}
			} else done(null);
		} catch (e) {
			done(null);
		}
	};
	xhr.onerror = function () {
		console.log("遥测错误 " + xhr.status);
		done(null);
	};
	xhr.open("POST", settings.url_telemetry + url_sep(settings.url_telemetry) + (settings.mpot ? "cors=true&" : "") + "r=" + Math.random(), true);
	const telemetryIspInfo = {
		processedString: clientIp,
		rawIspInfo: typeof ispInfo === "object" ? ispInfo : ""
	};
	try {
		const fd = new FormData();
		fd.append("ispinfo", JSON.stringify(telemetryIspInfo));
		fd.append("dl", dlStatus);
		fd.append("ul", ulStatus);
		fd.append("ping", pingStatus);
		fd.append("jitter", jitterStatus);
		fd.append("log", settings.telemetry_level > 1 ? log : "");
		fd.append("extra", settings.telemetry_extra);
		xhr.send(fd);
	} catch (ex) {
		const postData = "extra=" + encodeURIComponent(settings.telemetry_extra) + "&ispinfo=" + encodeURIComponent(JSON.stringify(telemetryIspInfo)) + "&dl=" + encodeURIComponent(dlStatus) + "&ul=" + encodeURIComponent(ulStatus) + "&ping=" + encodeURIComponent(pingStatus) + "&jitter=" + encodeURIComponent(jitterStatus) + "&log=" + encodeURIComponent(settings.telemetry_level > 1 ? log : "");
		xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
		xhr.send(postData);
	}
}