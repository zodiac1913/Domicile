const WAIT_OVERLAY_ID = "unobtrusiveWaitOverlay";

function getControllerName() {
	const pageContainer = document.querySelector("cc-container,sml-page");
	if (pageContainer?.dataset?.idPrefix) {
		return pageContainer.dataset.idPrefix;
	}

	const [, controller = ""] = location.pathname.split("/");
	return controller;
}

function showErrorMessage(message) {
	if (!message) {
		return;
	}

	if (typeof globalThis.modalBox === "function") {
		globalThis.modalBox(message, "Error", "danger");
		return;
	}

	console.warn(message);
}

export function clip(text, maxLength) {
	if (typeof text !== "string" || !text.trim()) {
		return "";
	}

	if (typeof maxLength !== "number" || maxLength < 1) {
		return text;
	}

	return text.length > maxLength ? text.slice(0, maxLength) : text;
}

export function asBool(value) {
	if (value === true) {
		return true;
	}

	if (value === false || value === null || value === undefined) {
		return false;
	}

	if (Array.isArray(value)) {
		return false;
	}

	if (typeof value === "string") {
		const normalizedValue = value.trim().toLowerCase();
		if (["true", "1", "y", "on"].includes(normalizedValue)) {
			return true;
		}

		return false;
	}

	if (typeof value === "number") {
		return value === 1;
	}

	return Boolean(value);
}

export function unAlterEncapse(str, encapseOne = "^~^", encapseOneReplace = '"') {
	if (!str) {
		return "{}";
	}

	return str.replaceAll("::ALTERENCAPSE::", "").replaceAll(encapseOne, encapseOneReplace);
}

export function asFieldNotationString(str) {
	if (!str || str.length < 1) {
		return "";
	}

	let normalizedString = str;
	if (normalizedString.includes(" ")) {
		normalizedString = normalizedString
			.split(" ")
			.map(segment => segment.substring(0, 1).toLowerCase() + segment.substring(1))
			.join("");
	}

	normalizedString = normalizedString.replaceAll(" ", "");
	if (normalizedString.length === 1) {
		return normalizedString.toLowerCase();
	}

	if (normalizedString.substring(0, 1) === normalizedString.substring(0, 1).toLowerCase()) {
		return `_${normalizedString}`;
	}

	return normalizedString.substring(0, 1).toLowerCase() + normalizedString.substring(1);
}

export function guid(noDash = false) {
	const uuid = typeof crypto?.randomUUID === "function"
		? crypto.randomUUID()
		: "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, character => {
				const randomValue = Math.random() * 16 | 0;
				const normalizedValue = character === "x" ? randomValue : (randomValue & 0x3 | 0x8);
				return normalizedValue.toString(16);
			});

	return noDash ? uuid.replaceAll("-", "") : uuid;
}

export function ensureStylesheetLink(linkId, href) {
	if (!linkId || !href || document.getElementById(linkId)) {
		return;
	}

	const head = document.head || document.getElementsByTagName("head")[0];
	if (!head) {
		return;
	}

	const link = document.createElement("link");
	link.id = linkId;
	link.rel = "stylesheet";
	link.type = "text/css";
	link.href = href;
	link.media = "all";
	head.appendChild(link);
}

export function isJson(str) {
	if (str === null || str === undefined) {
		return false;
	}

	if (typeof str === "object") {
		return true;
	}

	if (typeof str !== "string" || !str.trim()) {
		return false;
	}

	try {
		JSON.parse(str.trim());
		return true;
	} catch {
		return false;
	}
}

export function isJsonRepaired(str) {
	if (str === null || str === undefined) {
		return [false, str];
	}

	if (typeof str === "object") {
		return [true, str];
	}

	if (typeof str !== "string" || !str.trim()) {
		return [false, str];
	}

	const potentialJson = str.trim();
	try {
		JSON.parse(potentialJson);
		return [true, potentialJson];
	} catch (strictParseError) {
		console.log("Not valid JSON, attempting repair:", strictParseError.message);
		if (potentialJson.startsWith("{") && potentialJson.endsWith("}")) {
			let repaired = potentialJson;

			try {
				repaired = repaired.replaceAll("'", '"');
				repaired = repaired.replace(/([{,]\s*)(\w+)(\s*:)/g, '$1"$2"$3');
				JSON.parse(repaired);
				return [true, repaired];
			} catch (repairError) {
				console.log("Not valid JSON, failed repair", repairError);
				return [false, repaired];
			}
		}

		return [false, potentialJson];
	}
}

export function jmlToHtml(jsonIn, tabs = 2) {
	try {
		if (Array.isArray(jsonIn)) {
			return jsonIn.map(jObj => jmlToHtml(jObj, 1)).join("");
		}

		if (typeof jsonIn !== "object" || jsonIn === null) {
			return String(jsonIn);
		}

		const nodeType = (jsonIn.nodeType || jsonIn.node || jsonIn.n || "div").toLowerCase();
		const inputType = String(jsonIn.type || "").toLowerCase();
		const isNativeButtonControl =
			nodeType === "button" ||
			(nodeType === "input" && ["button", "submit", "reset", "image"].includes(inputType));
		let openTag = `<${nodeType}`;
		let children = "";

		const flatten1 = arr => {
			if (!Array.isArray(arr)) {
				return null;
			}

			return typeof arr.flat === "function" ? arr.flat(1) : [].concat(...arr);
		};

		const childArray = flatten1(jsonIn.b) || flatten1(jsonIn.babies);
		if (childArray) {
			children = childArray.map(obj => jmlToHtml(obj, tabs + 1)).join("");
		}

		const skipAttributeKeys = new Set([
			"nodeType", "node", "n",
			"b", "babies",
			"t", "text", "innerText", "innerHTML",
			"event", "e",
		]);
		const attributeMap = {
			c: "class",
			i: "id",
			v: "value",
			o: "onclick",
			s: "style",
			h: "href",
			r: "role",
			ttl: "title",
			p: "placeholder",
			acon: "aria-controls",
			aexp: "aria-expanded",
			ahdr: "aria-header",
			ahid: "aria-hidden",
			alab: "aria-label",
			albb: "aria-labelledby",
			alvl: "aria-level",
			areq: "aria-required",
			arol: "aria-role",
		};

		for (const key in jsonIn) {
			if (!Object.prototype.hasOwnProperty.call(jsonIn, key) || skipAttributeKeys.has(key)) {
				continue;
			}

			const value = jsonIn[key];
			if (typeof value === "object" && value !== null) {
				openTag += ` ${key}="[Object]"`;
				continue;
			}

			if (value === null || value === undefined) {
				openTag += ` ${key}="${String(value)}"`;
				continue;
			}

			const fullKey = attributeMap[key] || key;
			let safeValue = String(value).replace(/\"/g, "'");
			if (safeValue === "[object Object]") {
				safeValue = "Object";
			}

			if (fullKey === "role" && safeValue.toLowerCase() === "button" && isNativeButtonControl) {
				continue;
			}

			const quote = safeValue.includes('"') ? "'" : '"';
			openTag += ` ${fullKey}=${quote}${safeValue}${quote}`;

			if (fullKey === "title" && !jsonIn.hasOwnProperty("alab") && !jsonIn.hasOwnProperty("aria-label")) {
				openTag += ` aria-label=${quote}${safeValue}${quote}`;
			}
		}

		let innerText = "";
		if (jsonIn.hasOwnProperty("innerHTML")) {
			innerText = jsonIn.innerHTML || "";
		} else if (jsonIn.hasOwnProperty("t") || jsonIn.hasOwnProperty("text") || jsonIn.hasOwnProperty("innerText")) {
			innerText = jsonIn.t || jsonIn.text || jsonIn.innerText || "";
		}

		const endMarker = (jsonIn.hasOwnProperty("id") || jsonIn.hasOwnProperty("i"))
			? `<!-- data-end="${jsonIn.id || jsonIn.i}" -->`
			: "";

		const voidElements = new Set(["area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "source", "track", "wbr"]);

		openTag += ">";

		if (!(children && children.length) && !(innerText && innerText.length)) {
			if (voidElements.has(nodeType)) {
				return `${openTag.slice(0, -1)}>${endMarker}`;
			}

			return `${openTag}</${nodeType}>${endMarker}`;
		}

		return `${openTag}${innerText}${children}</${nodeType}>${endMarker}`;
	} catch (error) {
		console.log("[jmlToHtml] Exception:", error, jsonIn);
		return `<div class='jml-error'>Error rendering JML: ${error && error.message ? error.message : error}</div>`;
	}
}

export function toTitle(str) {
	if (!str || typeof str !== "string") {
		return "";
	}

	return str.toLowerCase().split(" ").map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(" ");
}

export function replaceLast(str, search, replacement, caseInsensitive = false) {
	if (search === "") {
		return str;
	}

	let modifiedString = str;
	let modifiedSearch = search;
	if (caseInsensitive) {
		modifiedString = str.toLowerCase();
		modifiedSearch = search.toLowerCase();
	}

	const lastIndex = modifiedString.lastIndexOf(modifiedSearch);
	if (lastIndex === -1) {
		return str;
	}

	const originalIndex = caseInsensitive ? str.indexOf(search, lastIndex) : lastIndex;
	return str.substring(0, originalIndex) + replacement + str.substring(originalIndex + search.length);
}

export function unobtrusiveWait(message = "Please wait...") {
	let waitDiv = document.getElementById(WAIT_OVERLAY_ID);
	if (!waitDiv) {
		waitDiv = document.createElement("div");
		waitDiv.id = WAIT_OVERLAY_ID;
		waitDiv.style.position = "fixed";
		waitDiv.style.top = "0";
		waitDiv.style.left = "0";
		waitDiv.style.width = "100vw";
		waitDiv.style.height = "100vh";
		waitDiv.style.background = "rgba(0,0,0,0.3)";
		waitDiv.style.display = "flex";
		waitDiv.style.alignItems = "center";
		waitDiv.style.justifyContent = "center";
		waitDiv.style.zIndex = "9999";
		waitDiv.style.fontSize = "2em";
		waitDiv.style.color = "#fff";
		document.body.appendChild(waitDiv);
	}

	waitDiv.textContent = message;
	waitDiv.role = "alert";
	waitDiv["aria-live"] = "assertive";
	waitDiv["aria-atomic"] = "true";
	waitDiv.style.display = "flex";
}

export function unobtrusiveWaitOff() {
	const waitDiv = document.getElementById(WAIT_OVERLAY_ID);
	if (waitDiv?.parentNode) {
		waitDiv.parentNode.removeChild(waitDiv);
	}
}

export async function apiPostDirect(url, dataUp, messageReceiverDataType = "json", noPost = false) {
	let requestUrl = url;
	if (requestUrl.indexOf("/") === -1) {
		const controller = getControllerName();
		requestUrl = `/${controller}/${requestUrl}`;
	}

	const requestBody = typeof dataUp === "object" ? JSON.stringify(dataUp) : dataUp;

	try {
		const response = await fetch(requestUrl, {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			cache: "no-cache",
			body: requestBody,
		});

		if (messageReceiverDataType === "text") {
			return await response.text();
		}

		const responseData = await response.json();
		if (!response.ok) {
			return {
				...responseData,
				errorObject: responseData?.errorObject || `Request failed with ${response.status} ${response.statusText}`,
			};
		}

		return responseData;
	} catch (error) {
		globalThis.dispatchEvent(new CustomEvent("pythia-server-unreachable"));
		return { errorObject: error?.message || "Request failed." };
	}
}

export async function receiptCheckGood(data) {
	if (typeof data === "string") {
		try {
			data = JSON.parse(data);
		} catch {
			return false;
		}
	}

	if (data?.Data && (data.Data.errorObject || data.Data.isValid !== undefined || data.Data.form)) {
		data = data.Data;
	}

	if (data?.errorObject) {
		unobtrusiveWaitOff();
		showErrorMessage(data.errorObject);
		return false;
	}

	if (data?.isValid === false) {
		unobtrusiveWaitOff();
		showErrorMessage(data.additionalInfo || "Validation failed.");
		return false;
	}

	return true;
}

export const smlUtilsReady = true;