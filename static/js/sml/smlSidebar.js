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
 * smlSidebar --- sml Page module for Encapsulating the page and ensuring a navbar and sidebar are present
 * Public Domain Licensed Copyright Law of the United States of America, Section 105 (https://www.copyright.gov/title17/92chap1.html#105)
 * Et qui me misit, mecum est: non reliquit me solum Pater, quia ego semper quae placita sunt ei, facio!
 * Published by: Dominic Roche of OIT/IUSG/DASM on 3/12/2026
 * @class smlSidebar
 * @extends {HTMLElement}
 */
// תהילתו. לא שלי
import sml from './sml.js';
import { clip, guid, jmlToHtml, toTitle } from "../smlUtils.js";
"use strict";
class smlSidebar extends HTMLElement {
  //----------------------Lifecycle

  constructor() {
    super();
    let ssb = this;
    ssb.innards="";
  }

    // connect component
  async connectedCallback() {
    let ssb = this;
    ssb.data=window.appSec;//get appSec
    //Build Sidebar Container
    if (!ssb.dataset.hasOwnProperty("opened")) ssb.dataset.opened = true;
    ssb.idPrefix =
      ssb?.dataset?.idPrefix || document.querySelector("sml-page")?.id || clip(guid(true),20);
    if (!ssb.id) ssb.id = ssb.idPrefix + "SMLSidebarAndPageContent";
    let sideBar=ssb.makeSidebarWrapper();
    //Build SideBar
    ssb.buildSidebar(sideBar);
    ssb.moveSlot();
    ssb.wire();
    window.pageSideBar = ssb;
  }

  static observedAttributes=["data-api"];


  //----------------------Lifecycle END

  //----------------------Render GUI

  buildSidebar(sideBar) {
    let ssb=this;
    ssb.insertAdjacentHTML("beforeend", jmlToHtml(sideBar));
    let sideBarGrouping = document.querySelector("#" + ssb.idPrefix + "PageSideBarRow");
    let sideBarState = getData("ccSidebarState");
    if (sideBarState) {
      if (sideBarState === "closed") {
        sideBarGrouping.classList.add("sml-sidebar-hidden");
      } else {
        sideBarGrouping.classList.remove("sml-sidebar-hidden");
      }
    } else {
      //use if no user setting
      if (!ssb.dataset.show) {
        sideBarGrouping.classList.remove("sml-sidebar-hidden");
      }
      else if (ssb.dataset.show === "false") {
        sideBarGrouping.classList.add("sml-sidebar-hidden");
      } else {
        sideBarGrouping.classList.remove("sml-sidebar-hidden");
      }
    }
    sideBarGrouping.classList.remove("d-none");
    //build user info
    ssb.buildUserInfo(sideBarGrouping);
    
    //build role buttons
    ssb.buildRoleButtons(sideBarGrouping);
    //bring it together
    let sidebarWrapper = document.querySelector("#" + ssb.dataset.idPrefix + "SidebarAndContentContainer");
    let pageContentWrapper = document.querySelector("#page-content-wrapper, #pageContentWrapper");
    if(!pageContentWrapper) { 
      pageContentWrapper={n: "div",i:"pageContentWrapper",c:"container-fluid p-0",b:[]};
      console.warn("Developer!  You must have a pageContentWrapper div as a reference of where the inner page begins.  Suggested html:" + 
        "&lt;div id='pageContentWrapper' class='container-fluid p-0' data-page-type='list'&gt;\n");
      let alertNoPageWrapper=makeAlertJML("alertNoPageWrapper","danger",
        "Developer!  You must have a pageContentWrapper div as a reference of where the inner page begins.  Suggested html:" + 
        "&lt;div id='pageContentWrapper' class='container-fluid p-0' data-page-type='list'&gt;\n");
      document.body.insertAdjacentHTML("afterbegin",jmlToHtml(alertNoPageWrapper));
      takeMeBack(document.querySelector("#alertNoPageWrapper"));
    }
    let pageContentWrapperContainer = document.querySelector("#" + ssb.idPrefix + "pageContent");
    pageContentWrapperContainer.insertAdjacentElement("afterBegin", pageContentWrapper);
  }



  makeSidebarWrapper() {
    let ssb = this;
    if (!ssb.dataset.hasOwnProperty("opened")) ssb.dataset.opened = true;
//        const sideBarContainer = {i:"sidebarAndContentContainer",c:"sml-sidebar-container d-flex",b:[]};

    let sbw = {
      i: ssb.idPrefix + "SidebarAndContentContainer",
      c:"sml-sidebar-wrap container-fluid m-0 p-0",// + (ssb.asBool(ssb.dataset.opened) ? "" : " d-none") this was for sb container old
      b: [],
    };
    sbw.b.push(ssb.makePageSidebar());

    return sbw;
  }

  makePageSidebar() {
    let ssb = this;
    let psb = {
      c: "sml-sidebar-shell d-flex",
      i: ssb.idPrefix + "PageSideBar",
      //"aria-expanded": "true",
      "data-prefix": ssb.idPrefix,
      b: [],
    };
    psb.b.push(ssb.makePageSidebarRow());
    psb.b.push({
      c: "sml-sidebar-page-content card flex-grow-1",
      i: ssb.idPrefix + "pageContent"})
    return psb;
  }

  makePageSidebarRow() {
    let ssb = this;
    let psr = { i: ssb.idPrefix + "PageSideBarRow",
      c: "sml-sidebar bg-dark bg-gradient text-white",
      "data-type": "sideBar", b: [] };
    return psr;
  }
  buildUserInfo(sideBarGrouping) {
    let userInfoSM={
      i:"usmall",c:"btn-group d-none",role:"group",ttl:"Current User Info", 
      alab:"Current User Info",b:[ 
        {c:"m-0 w-100 text-center text-center",b:
          [
            {n:"span",i:"userInfoSmall",c:"fs-5 ms-2 text-center pe-md-1 bi bi-person-circle border-0 text-white ",ttl:"User: Dominic F. Roche",alab:"User: Dominic F. Roche"},
          ]
        }
      ]
    };

    let userInfoMD = {
      n: "span", i: "uimedium", c: "fs-5 pe-md-1 border-0 text-white d-none sml-sidebar-user-medium",
      t: "",
      ttl: "User: " + toTitle(cso.CurrentEmployee.LegalFirstLastName), b: [
        { n: "br", c: "" },
        { n: "span", c: "ccsbButtonText ", t:  "" },
        { n: "span", c: "ccsbButtonText ", t: toTitle(cso.CurrentEmployee.LegalFirstLastName) }
      ]
    };

    let userInfoLG = {
      i: "uilgUserInfoDiv", c: "sml-sidebar-user-large", b: [
        { n: "span", i: "uilguserinfo", c: "ms-2 small text-white fst-italic", t: "User Information:" },
        { n: "br", c: "" },
        { n: "span", i: "uilguserinfoName", c: "ms-2 col-sm-12 fw-bold", t: toTitle(cso.CurrentEmployee.LegalFirstLastName) },
        { n: "br", c: "" },
        { n: "span", i: "uilguserinfoEchelon", c: "ms-2 col-sm-12 small", t: cso.CurrentEmployee.Echelon },
        { n: "br", c: "" }
      ]
    };
    sideBarGrouping.insertAdjacentHTML("afterbegin", jmlToHtml(userInfoSM));
    sideBarGrouping.insertAdjacentHTML("afterbegin", jmlToHtml(userInfoLG));
    sideBarGrouping.insertAdjacentHTML("afterbegin", jmlToHtml(userInfoMD));
  }

  buildRoleButtons(sideBarGrouping) {
    let ssb=this;
    let roleBtns = { i: ssb.idPrefix + "RoleButtonsGroup", c: "sml-sidebar-role-group btn-group-vertical d-flex flex-column mt-2", role: "group", ttl: "Menu of Roles for User in order of lowest to highest", b: [] };
    let rbBabies = {
      i: ssb.idPrefix + "RolesFauxButton",
      c: "sml-sidebar-group-label bg-responsive border-md-plus mb-1 font-weight-bold text-nowrap text-center w-100 p-1",
      b: []
    };

    if ((ssb.data?.rolesComponent?.roles || []).length > 0) {
      rbBabies.b.push({ i:"rolesHeaderBigScreen",n: "span", c: "", b:[
        {n:"span",c:"ccsbButtonText", t: "Roles" } ]});
      //rbBabies.b.push({ i:"rolesHeaderSmallScreen",n: "span", c: "showOnSmallScreen showOnXSScreen bi bi-grip-vertical" });
    }
    roleBtns.b.push(rbBabies);

    for (const rb of ssb.data.rolesComponent.roles) {
      if (rb.hasRole) {
        roleBtns.b.push({
          n: "sml-reactive-button", i: rb.elementId, type: "button", role: "button", ttl: "Assume " + rb.elementText + " role for " + rb.elementTitle,
          c: "btn " + (rb.isActive ? "btn-primary" : "btn-outline-primary") + " text-start"
            + " border-0 text-white mb-md-1 shadow-none text-nowrap text-truncate sidebarBtns sidebarRoleBtns sml-sidebar-btn", "data-active": rb.isActive,
          "data-rolename": rb.roleName,
          "data-elementtext": rb.elementText,
          "data-url":"",
          "data-api-mode":"change-role",
          "data-api": location.origin + "/" + cso.CurrentApp.ControllerName + "/ChangeRole/",
          "data-action": rb.action,
          "data-icon": rb.elementIcon,
          "data-text": rb.elementText,
        });
      }
    }
    sideBarGrouping.insertAdjacentHTML("beforeend", jmlToHtml(roleBtns));
  }




  //----------------------Render GUI END

  
  //----------------------Wire up
  wire(){
    let ssb=this;
    window.pagesmlSidebar=ssb;
    ssb.wireButtons();

  }

  wireButtons(){
    let ssb = this;
    let sidebarStyle=window.getData("ccSidebarStyle");

    //cool comms
    const setupSidebar = () => {
        if(sidebarStyle === "small"){
            pageNavBar.smallSidebarToggled(null, "small");

        } else {
            pageNavBar.smallSidebarToggled(null, "standard");
        }
    };
    
    if (typeof pageNavBar !== 'undefined' && pageNavBar.smallSidebarToggled) {
        setupSidebar();
    } else {
      window.addEventListener("navbarReady", setupSidebar, { once: true });
    }

    ssb.wireOtherButtons();

  }

  wireOtherButtons(){
    let ssb=this;
    let slotButtonsEles=Array.from(ssb.querySelectorAll(".sidebarBtns")).filter(b=> {
      if (b.classList.contains("sidebarRoleBtns")) return false;
      return (b.tagName || "").toLowerCase() !== "sml-reactive-button";
    });
    for(let b of slotButtonsEles){
      if(b.dataset.wired!=="true"){
        b.addEventListener("click",(e)=> ssb.buttonAction(e));
        b.dataset.wiredBy="wireOtherButtons";
        b.dataset.wired="true";
      }
    }
  }
  
  buttonAction(e){
    let btn=e.target.closest("button, sml-reactive-button");
    
    if(!btn?.dataset?.url){
      modalBox("No URL defined for this button. Please inform DASM","Button Error!");
    }else if(btn?.dataset?.url){
      if(btn.dataset.url.length<1){
        modalBox("No URL defined for this button. Please inform DASM","Button Error!");
      }else{
        if(btn.dataset.url===location.href){
          modalBox("You are already on the requested page.","Already There!");
        }else{
          this.primeButtonActiveUi(btn);
          if(btn.dataset.target === "_blank"){
            this.navigateAfterUiPaint(() => window.open(btn.dataset.url, '_blank'));
          }else{
            this.navigateAfterUiPaint(() => {
              location.href=btn.dataset.url;
            });
          }
        }
      }
    }
  }

  primeButtonActiveUi(btn) {
    if (!btn || (btn.tagName || "").toLowerCase() !== "sml-reactive-button") return;

    const canManageActive =
      btn.dataset.active !== undefined ||
      btn.dataset.toggles === "true" ||
      (typeof btn.shouldAutoActivateByUrl === "function" && btn.shouldAutoActivateByUrl());
    if (!canManageActive) return;

    if (typeof btn.shouldHandlePeers === "function" && btn.shouldHandlePeers()) {
      const peerGroup = typeof btn.resolvePeerGroup === "function" ? btn.resolvePeerGroup() : null;
      if (peerGroup && typeof btn.deactivatePeerButtonsInGroup === "function") {
        btn.deactivatePeerButtonsInGroup(peerGroup);
      }
    }

    btn.dataset.active = "true";
    if (typeof btn.applyActiveStyling === "function") btn.applyActiveStyling();
  }

  navigateAfterUiPaint(navigateFn) {
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        navigateFn();
      });
    });
  }

  

  moveSlot() {
    let ssb=this;
    if(ssb.innerHTML.length>0){
      let slotItems=Array.from(ssb.querySelectorAll("[data-part]"));
      let sbEle=document.querySelector("#" + ssb.idPrefix + "PageSideBarRow");
      if(sbEle){
        for(let el of slotItems){
          if(el.dataset && el.dataset.eid) el.id=el.dataset.eid;
          sbEle.insertAdjacentElement('beforeend', el);
        }
      }
      let slotdiv=ssb.querySelector("#slotdiv");
      if(slotdiv) slotdiv.remove();
    }
  }





  //----------------------Wire up




  takeOutTheGarbage(){
    let ssb=this;
    ssb.removeScriptByUrl("datatables");
    ssb.removeStyleByUrl("datatables");
  }

  removeScriptByUrl(keyword){
    document.querySelectorAll(`script[src*="${keyword}"]`)?.forEach(el => el.remove());
  }
  removeStyleByUrl(keyword){
    document.querySelectorAll(`link[href*="${keyword}"]`)?.forEach(el => el.remove());
  }


}

customElements.define("sml-sidebar", smlSidebar);
export default smlSidebar;
//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! S.D.G !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
//^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^