/*
	LibreSpeed - Worker
	作者：Federico Dossena
	https://github.com/librespeed/speedtest/
	GNU LGPLv3 许可证
*/

// --- 状态与数据 ---
let testState = -1; // -1=未开始, 0=启动中, 1=下载, 2=Ping, 3=上传, 4=完成, 5=中止
let dlStatus = ""; // 下载速度 (Mbps)
let ulStatus = ""; // 上传速度 (Mbps)
let pingStatus = ""; // Ping (ms)
let jitterStatus = ""; // Jitter (ms)
let clientIp = ""; // 客户端IP
let dlProgress = 0; // 下载进度 (0-1)
let ulProgress = 0; // 上传进度 (0-1)
let pingProgress = 0; // Ping进度 (0-1)
let testId = null; // 遥测测试ID

// --- 日志记录 ---
let log = "";
function tlog(s) { if (settings.telemetry_level >= 2) log += `${Date.now()}: ${s}\n`; }
function tverb(s) { if (settings.telemetry_level >= 3) log += `${Date.now()}: ${s}\n`; }
function twarn(s) {
	if (settings.telemetry_level >= 1) log += `${Date.now()} WARN: ${s}\n`;
	console.warn(s);
}

// --- 默认设置 (可被主线程覆盖) ---
let settings = {
	test_order: "IP_D_U",
	time_ul_max: 15,
	time_dl_max: 15,
	time_auto: true,
	time_ulGraceTime: 3,
	time_dlGraceTime: 1.5,
	count_ping: 10,
	url_dl: "backend/garbage",
	url_ul: "backend/empty",
	url_ping: "backend/empty",
	url_getIp: "backend/getIP",
	getIp_ispInfo: true,
	getIp_ispInfo_distance: "km",
	xhr_dlMultistream: 6,
	xhr_ulMultistream: 3,
	xhr_multistreamDelay: 300,
	xhr_ignoreErrors: 1, // 0=失败时中止, 1=重启流, 2=忽略错误
	xhr_dlUseBlob: false,
	xhr_ul_blob_megabytes: 20,
	garbagePhp_chunkSize: 100,
	enable_quirks: true,
	ping_allowPerformanceApi: true,
	overheadCompensationFactor: 1.06,
	useMebibits: false,
	telemetry_level: 0, // 默认值为0, 必须被主线程设置
	url_telemetry: "results/telemetry",
	telemetry_extra: "",
	forceIE11Workaround: false,
	mpot: false
};

let xhr = null; // 保存活动XHR请求的数组
let testInterval = null; // 用于测试中的setInterval
let testPointer = 0; // 指向settings.test_order中下一个测试的指针
let ispInfo = ""; // 用于遥测的ISP信息

function getUrlSeparator(url) { return url.includes('?') ? "&" : "?"; }

/**
 * 监听主线程消息
 */
self.addEventListener("message", (event) => {
	const { command, settings: newSettings } = event.data;

	if (command === "status") {
		postMessage({
			testState, dlStatus, ulStatus, pingStatus, clientIp, jitterStatus,
			dlProgress, ulProgress, pingProgress, testId
		});
		return;
	}

	if (command === "abort") {
		tlog("Test aborted by user");
		stopTest();
		testState = 5; // 设置为中止状态
		return;
	}

	if (command === "start" && testState === -1) {
		testState = 0;
		try {
			if (newSettings) {
				// 合并基础设置
				Object.assign(settings, newSettings);

				// --- 核心修复: 恢复遥测等级的字符串解析 ---
				if (typeof newSettings.telemetry_level === 'string') {
					settings.telemetry_level = { "basic": 1, "full": 2, "debug": 3 }[newSettings.telemetry_level] || 0;
				}
			}
			applyBrowserQuirks(newSettings || {});
			settings.test_order = settings.test_order.toUpperCase();
		} catch (e) {
			twarn(`Error parsing settings: ${e}`);
		}

		tverb(`Starting test with settings: ${JSON.stringify(settings)}`);
		testPointer = 0;
		runNextTest();
	}
});

function applyBrowserQuirks(userSettings) { /* ... 此函数保持不变 ... */
	if (!settings.enable_quirks) return;
	const ua = navigator.userAgent;

	if (/Firefox/i.test(ua) && userSettings.ping_allowPerformanceApi === undefined) {
		settings.ping_allowPerformanceApi = false;
	}
	if (/Edge/i.test(ua) && userSettings.xhr_dlMultistream === undefined) {
		settings.xhr_dlMultistream = 3;
	}
	if (/Chrome/i.test(ua) && self.fetch && userSettings.xhr_dlMultistream === undefined) {
		settings.xhr_dlMultistream = 5;
	}
	if (/Edge|PlayStation 4/i.test(ua)) {
		settings.forceIE11Workaround = true;
	}
	if (/Chrome/i.test(ua) && /Android|iPhone|iPad|iPod|Windows Phone/i.test(ua)) {
		settings.xhr_ul_blob_megabytes = 4;
	}
	if (/^((?!chrome|android|crios|fxios).)*safari/i.test(ua)) {
		settings.forceIE11Workaround = true;
	}
}

function clearRequests() { /* ... 此函数保持不变 ... */
	tverb("Clearing pending XHRs");
	if (!xhr) return;
	xhr.forEach(request => {
		if (request) {
			try { request.onprogress = request.onload = request.onerror = null; } catch (e) { }
			try { request.upload.onprogress = request.upload.onload = request.upload.onerror = null; } catch (e) { }
			try { request.abort(); } catch (e) { }
		}
	});
	xhr = null;
}

function stopTest() { /* ... 此函数保持不变 ... */
	clearRequests();
	if (testInterval) clearInterval(testInterval);
	if (settings.telemetry_level > 1) sendTelemetry(() => { });
	dlStatus = ulStatus = pingStatus = jitterStatus = "";
	dlProgress = ulProgress = pingProgress = 0;
}

/**
 * 按照test_order顺序执行下一个测试, 并在最后发送遥测数据
 */
function runNextTest() {
	if (testState === 5) return;

	if (testPointer >= settings.test_order.length) {
		// --- 核心逻辑: 所有测试完成, 准备发送遥测数据 ---
		tlog(`All tests finished. Telemetry level: ${settings.telemetry_level}.`);
		if (settings.telemetry_level > 0) {
			sendTelemetry(id => {
				testState = 4; // 报告已发送, 标记为完成
				if (id) testId = id;
			});
		} else {
			testState = 4; // 无需报告, 直接标记为完成
		}
		return;
	}

	const nextTest = settings.test_order.charAt(testPointer++);
	switch (nextTest) {
		case 'I': getIp(runNextTest); break;
		case 'D': testState = 1; dlTest(runNextTest); break;
		case 'U': testState = 3; ulTest(runNextTest); break;
		case 'P': testState = 2; pingTest(runNextTest); break;
		case '_': setTimeout(runNextTest, 1000); break;
		default: runNextTest();
	}
}

// getIp, dlTest, pingTest 函数保持不变, 这里省略以节约篇幅
function getIp(done) { /* ... 此函数保持不变 ... */
	tverb("getIp");
	const request = new XMLHttpRequest();
	request.withCredentials = true;
	request.onload = function () {
		try {
			const data = JSON.parse(this.responseText);
			clientIp = data.processedString;
			ispInfo = data.rawIspInfo;
		} catch (e) {
			clientIp = this.responseText;
			ispInfo = "";
		}
		done();
	};
	request.onerror = function () {
		twarn("getIp failed");
		done();
	};
	const url = `${settings.url_getIp}${getUrlSeparator(settings.url_getIp)}${settings.mpot ? "cors=true&" : ""}${settings.getIp_ispInfo ? `isp=true&distance=${settings.getIp_ispInfo_distance}&` : "&"}r=${Math.random()}`;
	request.open("GET", url, true);
	request.send();
}


/**
 * 下载测速 (基于用户提供的优秀实现集成)
 * @param {function} done 完成回调
 */
function dlTest(done) {
	tverb("dlTest: Starting download test...");
	testState = 1;

	let totalLoadedBytes = 0, startTime = Date.now(), bonusTime = 0, graceTimeDone = false, failed = false;
	xhr = [];

	const createStream = (streamIndex, delay) => {
		setTimeout(() => {
			if (testState !== 1) return;
			let prevLoaded = 0;
			const request = new XMLHttpRequest();
			xhr[streamIndex] = request;
			request.withCredentials = settings.mpot;
			request.onprogress = event => {
				if (testState !== 1) { try { request.abort(); } catch (e) {} return; }
				const diffLoaded = event.loaded - prevLoaded;
				if (!isNaN(diffLoaded) && isFinite(diffLoaded) && diffLoaded >= 0) {
					totalLoadedBytes += diffLoaded;
					prevLoaded = event.loaded;
				}
			};
			request.onload = () => {
				tverb(`dlTest: Stream ${streamIndex} finished, restarting.`);
				try { request.abort(); } catch (e) {}
				if (testState === 1) createStream(streamIndex, 0);
			};
			request.onerror = () => {
				twarn(`dlTest: Stream ${streamIndex} encountered an error.`);
				if (testState !== 1) return;
				if (settings.xhr_ignoreErrors === 0) failed = true;
				xhr[streamIndex] = null;
				if (settings.xhr_ignoreErrors === 1 && testState === 1) createStream(streamIndex, 0);
			};
			try { request.responseType = settings.xhr_dlUseBlob ? "blob" : "arraybuffer"; } catch (e) {}
			const url = new URL(settings.url_dl, self.location.origin);
			url.searchParams.append("r", Math.random());
			url.searchParams.append("ckSize", settings.garbagePhp_chunkSize);
			if (settings.mpot) url.searchParams.append("cors", "true");
			request.open("GET", url.toString(), true);
			request.send();
		}, 1 + delay);
	};

	tlog(`dlTest: Starting with ${settings.xhr_dlMultistream} streams.`);
	for (let i = 0; i < settings.xhr_dlMultistream; i++) {
		createStream(i, settings.xhr_multistreamDelay * i);
	}

	if (testInterval) clearInterval(testInterval);
	testInterval = setInterval(() => {
		const elapsedTime = Date.now() - startTime;
		if (!graceTimeDone) {
			dlProgress = elapsedTime / (settings.time_dlGraceTime * 1000);
			if (dlProgress > 1) dlProgress = 1;
			if (elapsedTime >= settings.time_dlGraceTime * 1000) {
				if (totalLoadedBytes > 0) {
					startTime = Date.now(); bonusTime = 0; totalLoadedBytes = 0;
				}
				graceTimeDone = true;
			}
		} else {
			const measurementTime = Date.now() - startTime;
			dlProgress = (measurementTime + bonusTime) / (settings.time_dl_max * 1000);
            if(dlProgress > 1) dlProgress = 1;

			const speedBps = totalLoadedBytes / (measurementTime / 1000.0);
			if (settings.time_auto) {
				const bonus = (5.0 * speedBps) / 100000;
				bonusTime += bonus > 400 ? 400 : bonus;
			}
			if (totalLoadedBytes > 0 && measurementTime > 100) {
				dlStatus = ((speedBps * 8 * settings.overheadCompensationFactor) / (settings.useMebibits ? 1048576 : 1000000)).toFixed(2);
			} else {
				dlStatus = "0.00";
			}

			if (((measurementTime + bonusTime) / 1000.0 > settings.time_dl_max) || failed) {
				if (failed || isNaN(parseFloat(dlStatus))) dlStatus = "Fail";
				clearRequests(); clearInterval(testInterval); testInterval = null;
				dlProgress = 1;
				tlog(`dlTest: Final download speed: ${dlStatus} Mbps.`);
				done();
				return;
			}
		}
	}, 200);
}

function pingTest(done) { /* ... 此函数保持不变 ... */
	tverb("pingTest");
	const startTime = Date.now();
	let prevTime = null, ping = 0.0, jitter = 0.0, receivedPongs = 0, prevInstSpeed = 0;
	xhr = [];

	const doPing = () => {
		pingProgress = receivedPongs / settings.count_ping;
		prevTime = Date.now();
		const request = new XMLHttpRequest();
		request.withCredentials = true;
		xhr[0] = request;

		request.onload = () => {
			let instSpeed;
			if (receivedPongs === 0) {
				instSpeed = Date.now() - prevTime;
			} else {
				instSpeed = Date.now() - prevTime;
				if (settings.ping_allowPerformanceApi) {
					try {
						const perfEntry = performance.getEntries().pop();
						const d = perfEntry.responseStart - perfEntry.requestStart;
						if (d > 0 && d < instSpeed) instSpeed = d;
					} catch (e) { }
				}
			}
			if (instSpeed < 1) instSpeed = prevInstSpeed || 1;

			const instJitter = Math.abs(instSpeed - prevInstSpeed);
			if (receivedPongs === 1) {
				ping = instSpeed;
				jitter = instJitter;
			} else {
				if (instSpeed < ping) ping = instSpeed;
				jitter = instJitter > jitter ? (jitter * 0.3 + instJitter * 0.7) : (jitter * 0.8 + instJitter * 0.2);
			}
			prevInstSpeed = instSpeed;

			pingStatus = ping.toFixed(2);
			jitterStatus = jitter.toFixed(2);
			receivedPongs++;
			tverb(`Ping: ${pingStatus} Jitter: ${jitterStatus}`);
			if (receivedPongs < settings.count_ping) doPing();
			else {
				pingProgress = 1;
				tlog(`pingTest result: ping ${pingStatus} jitter ${jitterStatus}`);
				done();
			}
		};

		request.onerror = () => {
			tverb("Ping failed");
			if (settings.xhr_ignoreErrors === 0) {
				pingStatus = "Fail";
				jitterStatus = "Fail";
				clearRequests();
				pingProgress = 1;
				done();
			} else {
				receivedPongs++;
				if (receivedPongs < settings.count_ping) doPing();
				else done();
			}
		};

		request.open("GET", `${settings.url_ping}${getUrlSeparator(settings.url_ping)}${settings.mpot ? "cors=true&" : ""}r=${Math.random()}`, true);
		request.send();
	};
	doPing();
}

/**
 * 上传测速 (已确认与遥测功能衔接正常)
 */
function ulTest(done) { /* ... 使用上一轮已修复的版本即可 ... */
	tverb("ulTest: Starting upload test...");
	let ulCalled = false; // 用于防止重复执行
	if (ulCalled) {
		tverb("ulTest: Upload test already attempted, skipping.");
		done();
		return;
	}
	ulCalled = true;
	testState = 3;

	const blobBaseChunk = new ArrayBuffer(1024 * 1024);
	try {
		const tempUint32Array = new Uint32Array(blobBaseChunk);
		for (let i = 0; i < tempUint32Array.length; i++) {
			tempUint32Array[i] = Math.random() * Math.pow(2, 32);
		}
	} catch (e) {
		twarn(`ulTest: Failed to create upload data template: ${e}`);
		ulStatus = "Fail"; ulProgress = 1; done(); return;
	}

	const uploadChunks = [];
	for (let i = 0; i < settings.xhr_ul_blob_megabytes; i++) {
		uploadChunks.push(blobBaseChunk);
	}
	const fullUploadBlob = new Blob(uploadChunks, { type: 'application/octet-stream' });
	const smallUploadBlob = new Blob([blobBaseChunk.slice(0, 256 * 1024)], { type: 'application/octet-stream' });

	const startUploadProcess = () => {
		tverb("ulTest: startUploadProcess function has been called");
		let totalUploadedBytes = 0.0, startTime = Date.now(), bonusTime = 0, graceTimeDone = false, failed = false;
		xhr = [];

		const createStream = (streamIndex, delay) => {
			setTimeout(() => {
				if (testState !== 3) return;
				let prevLoadedInStream = 0;
				const request = new XMLHttpRequest();
				xhr[streamIndex] = request;
				request.withCredentials = settings.mpot;

				let useIE11Workaround = settings.forceIE11Workaround;
				if (!useIE11Workaround) {
					try {
						if (!(request.upload && typeof request.upload.onprogress === 'object')) {
							useIE11Workaround = true;
							tlog(`ulTest: Stream ${streamIndex} - xhr.upload.onprogress not available, switching to IE11 workaround.`);
						}
					} catch (e) {
						useIE11Workaround = true;
						tlog(`ulTest: Stream ${streamIndex} - Error checking for onprogress (${e}), switching to IE11 workaround.`);
					}
				}

				const dataToSend = useIE11Workaround ? smallUploadBlob : fullUploadBlob;

				if (useIE11Workaround) {
					request.onload = () => {
						if (testState !== 3) { try { request.abort(); } catch (e) { } return; }
						totalUploadedBytes += smallUploadBlob.size;
						if (testState === 3) createStream(streamIndex, 0);
					};
					request.onerror = () => {
						if (testState !== 3) return;
						if (settings.xhr_ignoreErrors === 0) failed = true;
						xhr[streamIndex] = null;
						if (settings.xhr_ignoreErrors === 1 && testState === 3) createStream(streamIndex, 0);
					};
				} else {
					if (request.upload) {
						request.upload.onprogress = (event) => {
							if (testState !== 3) { try { request.abort(); } catch (e) { } return; }
							const diffLoaded = event.loaded > 0 ? event.loaded - prevLoadedInStream : 0;
							if (isNaN(diffLoaded) || !isFinite(diffLoaded) || diffLoaded < 0) return;
							totalUploadedBytes += diffLoaded;
							prevLoadedInStream = event.loaded;
						};
						request.upload.onload = () => {
							prevLoadedInStream = 0;
							if (testState === 3) createStream(streamIndex, 0);
						};
						request.upload.onerror = () => {
							if (testState !== 3) return;
							if (settings.xhr_ignoreErrors === 0) failed = true;
							xhr[streamIndex] = null;
							if (settings.xhr_ignoreErrors === 1 && testState === 3) createStream(streamIndex, 0);
						};
					} else {
						failed = true;
					}
				}

				const requestURL = new URL(settings.url_ul, self.location.origin);
				requestURL.searchParams.append("r", Math.random());
				if (settings.mpot) requestURL.searchParams.append("cors", "true");
				request.open("POST", requestURL.toString(), true);
				try { request.setRequestHeader("Content-Encoding", "identity"); } catch (e) { }
				try { request.send(dataToSend); } catch (e) {
					if (testState !== 3) return;
					if (settings.xhr_ignoreErrors === 0) failed = true;
					xhr[streamIndex] = null;
					if (settings.xhr_ignoreErrors === 1 && testState === 3) createStream(streamIndex, 0);
				}
			}, 1 + delay);
		};

		for (let i = 0; i < settings.xhr_ulMultistream; i++) {
			createStream(i, settings.xhr_multistreamDelay * i);
		}

		if (testInterval) clearInterval(testInterval);
		testInterval = setInterval(() => {
			const elapsedTime = Date.now() - startTime;
			if (!graceTimeDone) {
				ulProgress = elapsedTime / (settings.time_ulGraceTime * 1000);
				if (ulProgress > 1) ulProgress = 1;
				ulStatus = "0.00";
				if (elapsedTime >= settings.time_ulGraceTime * 1000) {
					if (totalUploadedBytes > 0) {
						startTime = Date.now();
						bonusTime = 0;
						totalUploadedBytes = 0.0;
					}
					graceTimeDone = true;
				}
			} else {
				const measurementTime = Date.now() - startTime;
				ulProgress = (measurementTime + bonusTime) / (settings.time_ul_max * 1000);
				if (ulProgress > 1) ulProgress = 1;
				const speedBps = totalUploadedBytes / (measurementTime / 1000.0);
				if (settings.time_auto) {
					const bonus = (5.0 * speedBps) / 100000;
					bonusTime += bonus > 400 ? 400 : bonus;
				}
				if (totalUploadedBytes > 0 && measurementTime > 100) {
					ulStatus = ((speedBps * 8 * settings.overheadCompensationFactor) / (settings.useMebibits ? 1048576 : 1000000)).toFixed(2);
				} else {
					ulStatus = "0.00";
				}
				if (((measurementTime + bonusTime) / 1000.0 > settings.time_ul_max) || failed) {
					if (failed || isNaN(parseFloat(ulStatus))) {
						ulStatus = "Fail";
					}
					clearRequests();
					clearInterval(testInterval);
					testInterval = null;
					ulProgress = 1;
					tlog(`ulTest: Final upload speed: ${ulStatus} Mbps.`);
					done();
					return;
				}
			}
		}, 200);
	};
	if (settings.mpot) {
		const preRequest = new XMLHttpRequest();
		preRequest.withCredentials = true;
		preRequest.onload = preRequest.onerror = () => { startUploadProcess(); };
		const requestURL = new URL(settings.url_ul, self.location.origin);
		requestURL.searchParams.append("cors", "true");
		preRequest.open("POST", requestURL.toString());
		preRequest.send();
	} else {
		startUploadProcess();
	}
}

/**
 * 发送遥测数据到后端
 */
function sendTelemetry(done) {
	tlog(`sendTelemetry: Preparing to send data. dl=${dlStatus}, ul=${ulStatus}, ping=${pingStatus}`);
	const request = new XMLHttpRequest();
	request.withCredentials = true;
	request.onload = () => {
		try {
			const [tag, id] = request.responseText.split(" ");
			done(tag === "id" ? id : null);
		} catch (e) { done(null); }
	};
	request.onerror = () => {
		twarn("Telemetry submission failed with status " + request.status);
		done(null);
	};
	request.open("POST", `${settings.url_telemetry}${getUrlSeparator(settings.url_telemetry)}${settings.mpot ? "cors=true&" : ""}`, true);

	const telemetryIspInfo = { processedString: clientIp, rawIspInfo: ispInfo || "" };
	try {
		const fd = new FormData();
		fd.append("ispinfo", JSON.stringify(telemetryIspInfo));
		fd.append("dl", dlStatus);
		fd.append("ul", ulStatus);
		fd.append("ping", pingStatus);
		fd.append("jitter", jitterStatus);
		if (settings.telemetry_level > 1) fd.append("log", log);
		fd.append("extra", settings.telemetry_extra);
		request.send(fd);
	} catch (ex) {
		const postData = `extra=${encodeURIComponent(settings.telemetry_extra)}&ispinfo=${encodeURIComponent(JSON.stringify(telemetryIspInfo))}&dl=${encodeURIComponent(dlStatus)}&ul=${encodeURIComponent(ulStatus)}&ping=${encodeURIComponent(pingStatus)}&jitter=${encodeURIComponent(jitterStatus)}&log=${encodeURIComponent(settings.telemetry_level > 1 ? log : "")}`;
		request.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
		request.send(postData);
	}
}