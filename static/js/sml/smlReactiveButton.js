//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! J.J. !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
//^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
//     *          |¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯|       †          _____          ↑
//   _____        |  o o o o o o  |      /|\        (     )         ↑
//  /  ^  \       | o o o o o o o |     / | \      (       )       / \
// /_/___\_\      |_______________|    /  |  \      (]¯¯¯[)       /   \
//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
/* eslint-disable no-undef */
/* eslint-disable no-console */
/*!
 * smlReactiveButton --- sml Reactive Button module for SML Reactive Buttons
 * Public Domain Licensed Copyright Law of the United States of America, Section 105 (https://www.copyright.gov/title17/92chap1.html#105)
 * Et qui me misit, mecum est: non reliquit me solum Pater, quia ego semper quae placita sunt ei, facio!
 * Published by: Dominic Roche of OIT/IUSG/DASM on 12/30/2025
 * @class smlReactiveButton
 * @extends {HTMLElement}
 */
// תהילתו. לא שלי
import "./sml.js";
import { apiPostDirect, clip, guid, jmlToHtml, receiptCheckGood, unobtrusiveWait, unobtrusiveWaitOff } from "../smlUtils.js";

class SmlReactiveButton extends HTMLElement {
  static observedAttributes=["data-api","data-toggler-done","data-active"];

    // connect component
  async connectedCallback() {
    let srb = this;
    srb.initializeBaseAttributes();
    srb.normalizeApiMetadata();
    srb.normalizeButtonTypeMetadata();
    srb.normalizeAuthorInnerText();
    srb.normalizeActiveMetadata();
    srb.applyButtonTypeIntentDefaults();
    srb.warnButtonTypeConflicts();
    srb.applyUrlActiveState();
    if (srb.hasManagedActiveState()) {
      srb.setAttribute("aria-pressed", srb.dataset.active === "true");
    } else {
      srb.removeAttribute("aria-pressed");
    }
    srb.dataset.icon= srb.dataset.icon || "bi bi-question-circle";
    if(!srb.dataset.text) {
      srb.dataset.text = srb.textContent.trim() || "Whats this do?";
    }
    srb.buildButtonMarkup();
    srb.ensureBaseClasses();
    srb.wire();
    srb.applyActiveStyling();
  }

  normalizeAuthorInnerText() {
    let srb = this;
    if (srb.dataset.authorTextNormalized === "true") return;

    const authoredInnerText = Array.from(srb.childNodes)
      .filter(node => node.nodeType === Node.TEXT_NODE)
      .map(node => (node.textContent || "").trim())
      .filter(text => text.length > 0)
      .join(" ")
      .trim();
    if (!authoredInnerText) return;

    const existingDataText = (srb.dataset.text || "").trim();
    if (!existingDataText) {
      srb.dataset.text = authoredInnerText;
      srb.dataset.authorTextNormalized = "true";
      return;
    }

    if (existingDataText === authoredInnerText) {
      srb.dataset.authorTextNormalized = "true";
      return;
    }

    const combinedText = (existingDataText + " " + authoredInnerText)
      .replaceAll(/\s+/g, " ")
      .trim();
    srb.dataset.text = combinedText;
    srb.dataset.authorTextNormalized = "true";
  }

  initializeBaseAttributes() {
    let srb = this;
    srb.id = srb.id || "SRB"+clip(guid(true),20);
    if(!srb.type) srb.type="button";
    if(!srb.role) srb.role="button";
    if(!srb.hasAttribute("tabindex")) srb.tabIndex = 0;
    if(!srb.title) srb.title="sml Reactive Button not explained, ask DASM to fix this";
    if(!srb.ariaLabel) srb.ariaLabel=srb.title;
    if(srb.className=="") srb.className="smlRB btn btn-outline-primary border-0 text-white text-nowrap text-truncate";
  }

  normalizeApiMetadata() {
    let srb = this;
    srb.dataset.api = srb.dataset.api || "";
    if(srb.dataset.api && !srb.dataset.apiForward) {
      srb.dataset.apiForward = srb.dataset.api.replaceAll("/","_").replaceAll("?","_").replaceAll("&","_").replaceAll("=","_");
      console.log("smlReactiveButton: Setting apiForward to "+srb.dataset.apiForward);
    }
  }

  normalizeActiveMetadata() {
    let srb = this;
    const hasDataActive = srb.dataset.active !== undefined;
    if (!hasDataActive) {
      delete srb.dataset.active;
      return;
    }
    const isActive = (srb.dataset.active || "false").toString().toLowerCase() === "true";
    srb.dataset.active = isActive ? "true" : "false";
  }

  getSupportedButtonTypes() {
    return ["stateless", "toggle", "link", "command", "menu", "destructive"];
  }

  inferButtonType() {
    let srb = this;
    if ((srb.dataset.apiMode || "").toLowerCase() === "change-role") return "command";
    if ((srb.dataset.apiMode || "").toLowerCase() === "table-action") return "command";
    if (srb.dataset.handlePeers === "true" || srb.dataset.active !== undefined || srb.dataset.toggles === "true") return "toggle";
    if ((srb.dataset.url || "") !== "" && (srb.dataset.url || "") !== "null") return "link";
    if ((srb.dataset.api || "") !== "") return "command";
    if (srb.dataset.menu === "true" || srb.getAttribute("aria-haspopup") === "true") return "menu";
    if (srb.classList.contains("btn-danger") || srb.classList.contains("btn-outline-danger")) return "destructive";
    return "stateless";
  }

  normalizeButtonTypeMetadata() {
    let srb = this;
    const supported = srb.getSupportedButtonTypes();
    const rawType = (srb.dataset.buttonType || "").toString().trim().toLowerCase();
    const normalizedType = rawType || srb.inferButtonType();

    if (rawType && !supported.includes(rawType)) {
      if (srb.dataset.buttonTypeWarned !== "true") {
        console.warn("smlReactiveButton: Unsupported data-button-type='" + rawType + "'. Falling back to inferred type.", srb.id || srb);
        srb.dataset.buttonTypeWarned = "true";
      }
      srb.dataset.buttonType = srb.inferButtonType();
      return;
    }

    srb.dataset.buttonType = normalizedType;
  }

  applyButtonTypeIntentDefaults() {
    let srb = this;
    const buttonType = (srb.dataset.buttonType || "stateless").toLowerCase();
    if (buttonType === "toggle" && srb.dataset.active === undefined && srb.dataset.toggles === undefined) {
      srb.dataset.toggles = "true";
    }
    if (buttonType === "menu" && !srb.hasAttribute("aria-haspopup")) {
      srb.setAttribute("aria-haspopup", "true");
    }
  }

  warnButtonTypeConflicts() {
    let srb = this;
    if (srb.dataset.buttonTypeConflictWarned === "true") return;

    const buttonType = (srb.dataset.buttonType || "stateless").toLowerCase();
    const hasUrl = !!(srb.dataset.url && srb.dataset.url !== "" && srb.dataset.url !== "null");
    const hasApi = !!(srb.dataset.api && srb.dataset.api !== "");
    const hasManagedToggle = srb.dataset.active !== undefined || srb.dataset.toggles === "true";
    const apiMode = (srb.dataset.apiMode || "").toLowerCase();

    if (buttonType === "link" && !hasUrl) {
      console.warn("smlReactiveButton: data-button-type='link' without data-url.", srb.id || srb);
    }
    if (buttonType === "toggle" && !hasManagedToggle) {
      console.warn("smlReactiveButton: data-button-type='toggle' has no managed active metadata.", srb.id || srb);
    }
    if (buttonType === "stateless" && (hasUrl || hasApi || hasManagedToggle || apiMode === "change-role")) {
      console.warn("smlReactiveButton: data-button-type='stateless' conflicts with interactive action/state metadata.", srb.id || srb);
    }
    if (buttonType === "menu" && (hasUrl || hasApi || apiMode === "change-role")) {
      console.warn("smlReactiveButton: data-button-type='menu' should not mix with API/URL action wiring.", srb.id || srb);
    }

    srb.dataset.buttonTypeConflictWarned = "true";
  }

  buildButtonMarkup() {
    let srb = this;
    const isIconOnly = srb.dataset.iconOnly === "true" || srb.classList.contains("sml-icon-only");
    //build reactive button
    srb.innerHTML='';
    //outer span
    let outerSpan = {n:"span",c: srb.dataset.icon + " p-0",
      b:[{ n: "span", c: "sr-only visually-hidden", s:"position:absolute;width:1px;height:1px;padding:0;margin:-1px;overflow:hidden;clip:rect(0,0,0,0);white-space:nowrap;border:0;", t: srb.dataset.text }]};
    //inner span (omit for icon-only button variants)
    if(!isIconOnly){
      let innerSpan = {n:"span",c:"ms-2 smlRBText hideOnMediumOrLessWindow",t:srb.dataset.text};
      outerSpan.b.push(innerSpan);
    }
    srb.insertAdjacentHTML("afterbegin",jmlToHtml(outerSpan));
  }

  ensureBaseClasses() {
    let srb = this;
    if(!srb.classList.contains("smlRB")) srb.classList.add("smlRB");
    if(!srb.classList.contains("text-truncate")) srb.classList.add("text-truncate");
    if(!srb.classList.contains("text-nowrap")) srb.classList.add("text-nowrap");
  }

  // attribute change handler
  /**
   * This updates class properties based on html tags being altered
   *
   * @param {*} name name of element tag
   * @param {*} oldValue old value for reference of element tag
   * @param {*} newValue new value of element tag
   * @memberof smlReactiveButton
   */
  async attributeChangedCallback(name, oldValue, newValue) {
      let srb=this;
      if (oldValue === newValue) return;
      switch(name){
              case "data-api":
            srb.api=newValue || "";
                  break;
            case "data-active":
          if (newValue === null || newValue === undefined) {
            delete srb.dataset.active;
          } else {
            srb.dataset.active = (newValue + "").toLowerCase() === "true" ? "true" : "false";
          }
              srb.applyActiveStyling();
              break;
              default: //WATDUH!
                  console.log(`smlReactiveButton: What you talking bout Willis? for attribute change ${name} from ${oldValue} to ${newValue}`);
                  break;
      }
  }

  wire() {
      let srb = this;
      const wasPreviouslyWired = srb.dataset.wired === "true";
      const isManualWire = srb.dataset.manualWire === "true";
      const skipApiWiring = wasPreviouslyWired || isManualWire;
      const hasApi = !!(srb.dataset.api && srb.dataset.api !== "");
      const hasUrl = !!(srb.dataset.url && srb.dataset.url !== "" && srb.dataset.url !== "null");

      //For API Calls
      if(!skipApiWiring && hasApi && !hasUrl && srb.dataset.apiWired!=="true") {
          srb.addEventListener("click", async (e) => { srb.apiCall(e);});
          srb.dataset.apiWired="true";
      }
      //For Url (Navigation)
      setTimeout(()=>{unobtrusiveWaitOff();},10000);
      if(hasUrl && srb.dataset.urlWired!=="true") {
          srb.addEventListener("click", async (e) => {
            if (hasApi) {
              await srb.apiCall(e);
            }
              unobtrusiveWait("Going to " + e.currentTarget.dataset.url);
            globalThis.location.href=srb.dataset.url;
          });
          if (hasApi) srb.dataset.apiWired="true";
          srb.dataset.urlWired="true";
      }else{
          console.log("Not wiring url. url:" + srb.dataset.url + " wired:" + srb.dataset.wired);
      }
      srb.tabIndex = 0;

      if(srb.dataset.keyWired!=="true"){
        srb.addEventListener("keydown", (e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            srb.click();
          }
        });
        srb.dataset.keyWired="true";
      }

      if(srb.dataset.activeClickWired!=="true") {
        srb.addEventListener("click", () => {
          srb.toggleActiveStateOnClick();
        });
        srb.dataset.activeClickWired = "true";
      }

      srb.applyUrlActiveState();
      srb.applyActiveStyling();
  srb.dataset.wired = "true";
      
  }

  shouldAutoActivateByUrl() {
    let srb = this;
    if (!srb.dataset.url || srb.dataset.url === "" || srb.dataset.url === "null") return false;
    if (srb.dataset.activeByUrl === "false") return false;
    if (srb.dataset.activeByUrl === "true") return true;
    return srb.classList.contains("sidebarRelatedBtns");
  }

  normalizePath(pathValue) {
    let path = (pathValue || "/").toString().trim();
    if (path.length < 1) path = "/";
    path = path.replace(/\/+$/, "");
    if (path.length < 1) path = "/";
    return path.toLowerCase();
  }

  normalizeQueryParamKey(paramKey) {
    return (paramKey || "").toString().trim().toLowerCase();
  }

  normalizeQueryParamValue(paramValue) {
    return (paramValue ?? "").toString().trim();
  }

  collectQueryParams(searchParams) {
    let paramsMap = new Map();
    searchParams.forEach((value, key) => {
      const normalizedKey = this.normalizeQueryParamKey(key);
      const normalizedValue = this.normalizeQueryParamValue(value);
      if (!paramsMap.has(normalizedKey)) paramsMap.set(normalizedKey, []);
      paramsMap.get(normalizedKey).push(normalizedValue);
    });
    return paramsMap;
  }

  hasQuerySubset(targetParams, currentParams) {
    if (targetParams.size === 0) return true;

    for (let [key, targetValues] of targetParams.entries()) {
      const currentValues = currentParams.get(key) || [];
      if (currentValues.length < targetValues.length) return false;

      let counts = new Map();
      for (let value of currentValues) {
        counts.set(value, (counts.get(value) || 0) + 1);
      }

      for (let value of targetValues) {
        const remaining = counts.get(value) || 0;
        if (remaining < 1) return false;
        counts.set(value, remaining - 1);
      }
    }

    return true;
  }

  applyUrlActiveState() {
    let srb = this;
    if (!srb.shouldAutoActivateByUrl()) return;

    try {
      const targetUrl = new URL(srb.dataset.url, location.origin);
      const currentUrl = new URL(location.href);
      const targetPath = srb.normalizePath(targetUrl.pathname);
      const currentPath = srb.normalizePath(currentUrl.pathname);

      if (targetPath !== currentPath) {
        srb.dataset.active = "false";
        return;
      }

      const targetQueryParams = srb.collectQueryParams(targetUrl.searchParams);
      if (targetQueryParams.size === 0) {
        srb.dataset.active = "true";
        return;
      }

      const currentQueryParams = srb.collectQueryParams(currentUrl.searchParams);
      srb.dataset.active = srb.hasQuerySubset(targetQueryParams, currentQueryParams)
        ? "true"
        : "false";
    } catch {
      srb.dataset.active = "false";
    }
  }

  applyActiveStyling() {
    let srb = this;
    if (!srb.hasManagedActiveState()) {
      srb.style.removeProperty("background-color");
      srb.style.removeProperty("border-color");
      srb.style.removeProperty("box-shadow");
      srb.style.removeProperty("color");
      srb.removeAttribute("aria-current");
      srb.removeAttribute("aria-pressed");
      return;
    }

    const isActive = (srb.dataset.active || "false").toLowerCase() === "true";
    srb.setAttribute("aria-pressed", isActive ? "true" : "false");

    if (!isActive) {
      srb.style.removeProperty("background-color");
      srb.style.removeProperty("border-color");
      srb.style.removeProperty("box-shadow");
      srb.style.removeProperty("color");
      srb.removeAttribute("aria-current");
      return;
    }

    // Active UI is default-on for managed active buttons; set data-active-style-auto="false" to keep ARIA-only state.
    if (srb.dataset.activeStyleAuto === "false") {
      srb.setAttribute("aria-current", "page");
      return;
    }

    const minContrast = srb.resolveContrastTarget();
    const textColor = srb.resolveCssColor(globalThis.getComputedStyle(srb).color || "#ffffff");
    const inverseColor = srb.invertColor(textColor);
    const activeBg = srb.findClosestContrastColor(inverseColor, textColor, minContrast);
    const borderSeed = srb.mixColor(activeBg, textColor, 0.18);
    const borderColor = srb.findClosestContrastColor(borderSeed, textColor, 3);

    srb.style.setProperty("background-color", srb.rgbToCss(activeBg), "important");
    srb.style.setProperty("border-color", srb.rgbToCss(borderColor), "important");
    srb.style.setProperty("color", srb.rgbToCss(textColor), "important");
    srb.style.setProperty("box-shadow", "inset 0 1px 0 rgba(255, 255, 255, 0.22), 0 0 0 1px rgba(255, 255, 255, 0.10)", "important");
    srb.setAttribute("aria-current", "page");
  }

  hasManagedActiveState() {
    let srb = this;
    return srb.dataset.active !== undefined || srb.dataset.toggles === "true" || srb.shouldAutoActivateByUrl();
  }

  shouldHandlePeers() {
    let srb = this;
    if (srb.dataset.handlePeers === "true") return true;
    if (srb.dataset.handlePeers === "false") return false;
    return srb.classList.contains("sidebarRelatedBtns");
  }

  toggleActiveStateOnClick() {
    let srb = this;
    const apiMode = (srb.dataset.apiMode || "").toLowerCase();
    if (!srb.hasManagedActiveState()) return;
    if (apiMode === "change-role") return;

    const wasActive = (srb.dataset.active || "false").toLowerCase() === "true";
    if (srb.shouldHandlePeers()) {
      srb.handlePeerManagedClick(wasActive);
    } else {
      srb.dataset.active = wasActive ? "false" : "true";
    }

    srb.applyActiveStyling();
  }

  handlePeerManagedClick(wasActive) {
    let srb = this;
    const peerGroup = srb.resolvePeerGroup();
    if (!peerGroup) {
      if (srb.dataset.handlePeers === "true") srb.warnMissingPeerGroup();
      srb.dataset.active = wasActive ? "false" : "true";
      return;
    }

    // Highlander mode: exactly one active in the peer group.
    if (wasActive) {
      srb.dataset.active = "true";
      return;
    }

    srb.deactivatePeerButtonsInGroup(peerGroup);
    srb.dataset.active = "true";
  }

  resolvePeerGroup() {
    let srb = this;
    return srb.closest(".btn-group-vertical, .btn-group, [role='group']");
  }

  warnMissingPeerGroup() {
    let srb = this;
    if (srb.dataset.peerGroupWarned === "true") return;
    console.warn("smlReactiveButton: data-handle-peers='true' but no nearest .btn-group/.btn-group-vertical/[role='group'] was found.", srb.id || srb);
    srb.dataset.peerGroupWarned = "true";
  }

  deactivatePeerButtonsInGroup(parentGroup) {
    let srb = this;
    if (!parentGroup) return;

    const siblings = Array.from(parentGroup.querySelectorAll("sml-reactive-button"));
    for (let sibling of siblings) {
      if (sibling === srb) continue;
      if (!sibling.hasManagedActiveState?.()) continue;
      sibling.dataset.active = "false";
      if (typeof sibling.applyActiveStyling === "function") sibling.applyActiveStyling();
    }
  }

  resolveContrastTarget(){
    // Default is Section 508 / WCAG AA contrast for normal text.
    const defaultContrast=4.5;
    const aaaContrast=7;
    const parentSidebar = this.closest("sml-sidebar");

    const buttonMin=Number.parseFloat(this.dataset?.contrastMin || "");
    if(Number.isFinite(buttonMin)) return this.clampContrastValue(buttonMin);

    const sidebarMin=Number.parseFloat(parentSidebar?.dataset?.contrastMin || "");
    if(Number.isFinite(sidebarMin)) return this.clampContrastValue(sidebarMin);

    const preference=((this.dataset?.contrastStandard || parentSidebar?.dataset?.contrastStandard || "") + "")
      .trim()
      .toUpperCase();
    if(preference==="AAA") return aaaContrast;
    if(preference==="AA" || preference==="508" || preference==="SECTION508" || preference==="WCAG-AA") {
      return defaultContrast;
    }

    return defaultContrast;
  }

  clampContrastValue(value){
    return Math.max(1, Math.min(21, value));
  }

  resolveCssColor(colorValue){
    let srb=this;
    let parsed=srb.parseColorValue(colorValue);
    if(parsed) return parsed;

    let probe=document.createElement("span");
    probe.style.color=colorValue;
    probe.style.display="none";
    document.body.appendChild(probe);
    let resolved=globalThis.getComputedStyle(probe).color;
    probe.remove();
    return srb.parseColorValue(resolved)||{r:255,g:255,b:255};
  }

  parseColorValue(colorValue){
    if(!colorValue || typeof colorValue!=="string") return null;
    const rgbRegex=/^rgba?\(([^)]+)\)$/i;
    const rgbMatch=rgbRegex.exec(colorValue.trim());
    if(rgbMatch){
      let parts=rgbMatch[1].split(",").map(p=>Number.parseFloat(p.trim()));
      if(parts.length>=3){
        return {
          r: Math.max(0, Math.min(255, Math.round(parts[0]))),
          g: Math.max(0, Math.min(255, Math.round(parts[1]))),
          b: Math.max(0, Math.min(255, Math.round(parts[2])))
        };
      }
    }

    if(colorValue.startsWith("#")){
      let hex=colorValue.slice(1).trim();
      if(hex.length===3){
        hex=hex.split("").map(h=>h+h).join("");
      }
      if(hex.length===6){
        return {
          r: Number.parseInt(hex.slice(0,2),16),
          g: Number.parseInt(hex.slice(2,4),16),
          b: Number.parseInt(hex.slice(4,6),16)
        };
      }
    }
    return null;
  }

  invertColor(color){
    return {
      r: 255-color.r,
      g: 255-color.g,
      b: 255-color.b
    };
  }

  mixColor(a,b,t){
    return {
      r: Math.round(a.r + (b.r-a.r)*t),
      g: Math.round(a.g + (b.g-a.g)*t),
      b: Math.round(a.b + (b.b-a.b)*t),
    };
  }

  rgbToCss(color){
    return `rgb(${color.r}, ${color.g}, ${color.b})`;
  }

  channelToLinear(channel){
    let c=channel/255;
    return c<=0.03928 ? c/12.92 : Math.pow((c+0.055)/1.055, 2.4);
  }

  relativeLuminance(color){
    return (0.2126* this.channelToLinear(color.r))
      + (0.7152* this.channelToLinear(color.g))
      + (0.0722* this.channelToLinear(color.b));
  }

  contrastRatio(a,b){
    let l1=this.relativeLuminance(a);
    let l2=this.relativeLuminance(b);
    let high=Math.max(l1,l2);
    let low=Math.min(l1,l2);
    return (high+0.05)/(low+0.05);
  }

  colorDistance(a,b){
    let dr=a.r-b.r;
    let dg=a.g-b.g;
    let db=a.b-b.b;
    return (dr*dr)+(dg*dg)+(db*db);
  }

  findClosestContrastColor(startColor, textColor, minContrast){
    if(this.contrastRatio(startColor, textColor)>=minContrast) return startColor;

    const black={r:0,g:0,b:0};
    const white={r:255,g:255,b:255};
    const towardBlack=this.searchTowardsContrast(startColor, black, textColor, minContrast);
    const towardWhite=this.searchTowardsContrast(startColor, white, textColor, minContrast);

    if(towardBlack && towardWhite){
      return this.colorDistance(startColor, towardBlack) <= this.colorDistance(startColor, towardWhite)
        ? towardBlack
        : towardWhite;
    }
    if(towardBlack) return towardBlack;
    if(towardWhite) return towardWhite;

    return this.contrastRatio(black, textColor) >= this.contrastRatio(white, textColor)
      ? black
      : white;
  }

  searchTowardsContrast(startColor, targetColor, textColor, minContrast){
    if(this.contrastRatio(targetColor, textColor)<minContrast) return null;

    let low=0;
    let high=1;
    let best=null;
    for(let i=0;i<24;i++){
      let mid=(low+high)/2;
      let candidate=this.mixColor(startColor, targetColor, mid);
      if(this.contrastRatio(candidate, textColor)>=minContrast){
        best=candidate;
        high=mid;
      }else{
        low=mid;
      }
    }
    return best;
  }

  async apiCall() {
    let srb = this;
    if ((srb.dataset.apiMode || "").toLowerCase() === "change-role") {
      await srb.apiCallChangeRole();
      return;
    }
    if ((srb.dataset.apiMode || "").toLowerCase() === "table-action") {
      await srb.apiCallTableAction();
      return;
    }
    //api call logic here
    unobtrusiveWait("Processing " + srb.dataset.text);
    let postData=await apiPostDirect(srb.dataset.api, { forward: srb.dataset.apiForward });
    unobtrusiveWaitOff();
    const hasValidReceipt = postData?.receipt ? await receiptCheckGood(postData.receipt) : false;
    if (hasValidReceipt) {
        //success logic here
      globalThis[srb.dataset.apiForward](postData);
    } else {
        //failure logic here
        console.log("smlReactiveButton: API Call failed for "+srb.dataset.api);
    }
    
  }

  async apiCallTableAction() {
    let srb = this;
    const tableId = srb.dataset.table || "";
    const ownerTable = srb.closest("cc-table") || globalThis[tableId];

    if (!ownerTable || typeof ownerTable.handleActionButton !== "function") {
      console.warn("smlReactiveButton: table-action mode could not find owning cc-table.", srb.id || srb);
      return;
    }

    await ownerTable.handleActionButton(srb, { currentTarget: srb, target: srb });
  }

  async apiCallChangeRole() {
    let srb = this;
    const appSecurity = globalThis.appSec;
    const appContext = globalThis.cso;
    const showModal = globalThis.modalBox;
    let newRole = srb.dataset.rolename;
    let newElementText = srb.dataset.elementtext || srb.dataset.text || newRole;
    const controllerName = appContext?.CurrentApp?.ControllerName;

    if (!newRole) {
      console.warn("smlReactiveButton.changeRole: Missing data-rolename.");
      return;
    }

    if (newRole === appSecurity?.currentRole) {
      if (typeof showModal === "function") {
        showModal(
          "You are already configured for " + newElementText + " role!",
          "Already in Role"
        );
      }
      return;
    }

    const dataUp = {
      roleName: newRole,
      userIdentifier: appContext?.UserIdentifier,
      controllerAction: "ChangeRole",
      controllerName: controllerName,
    };

    unobtrusiveWait("Changing your role to " + newElementText + ". Please wait!",
      srb.id + "changeRole", srb.id);
    const api = srb.dataset.api || ("/" + controllerName + "/ChangeRole/");
    const data = await apiPostDirect(api, JSON.stringify(dataUp), "json", true);
    unobtrusiveWaitOff();

    if (!(await receiptCheckGood(data)) || data?.hasOwnProperty("redirect")) return;

    if (data?.hasOwnProperty("errorObject")) {
      console.warn("smlReactiveButton.changeRole: API returned errorObject.", data.errorObject);
      return;
    }

    const roleButtonGroup = srb.closest(".btn-group-vertical");
    const buttons = Array.from((roleButtonGroup || document).querySelectorAll(".sidebarRoleBtns"));
    const oldButton = buttons.find(b => b.classList.contains("btn-primary"));
    if (oldButton && oldButton !== srb) {
      oldButton.classList.remove("btn-primary");
      oldButton.classList.add("btn-outline-primary");
      oldButton.dataset.active = "false";
    }

    srb.classList.remove("btn-outline-primary");
    srb.classList.add("btn-primary");
    srb.dataset.active = "true";
    if (appSecurity) appSecurity.currentRole = newRole;

    const action = srb.dataset.action || data?.rolesComponent?.roles?.find(r => r.roleName === newRole)?.action;
    const dataControllerName = data?.currentApp?.ControllerName || controllerName;
    if (!action || !dataControllerName) {
      location.reload();
      return;
    }

    unobtrusiveWait("Changing your role to " + newRole + ". Please wait!",
      srb.id + "changeRole", srb.id);
    const hostRoot = (typeof globalThis.host === "string" && globalThis.host.length > 0) ? globalThis.host : location.origin;
    location.href = hostRoot + "/" + dataControllerName + "/" + action;
  }




}

customElements.define("sml-reactive-button", SmlReactiveButton);

export default SmlReactiveButton;

//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! S.D.G !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
//^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^