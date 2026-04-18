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
 * smlValid --- sml FormField Validation module for SML FormFields
 * Public Domain Licensed Copyright Law of the United States of America, Section 105 (https://www.copyright.gov/title17/92chap1.html#105)
 * Et qui me misit, mecum est: non reliquit me solum Pater, quia ego semper quae placita sunt ei, facio!
 * Published by: Dominic Roche of OIT/IUSG/DASM on 1/5/2026
 * @class smlValid
 * @extends {HTMLElement}
 */
// תהילתו. לא שלי
import { clip, guid, jmlToHtml, unobtrusiveWait, unobtrusiveWaitOff } from "../smlUtils.js";
"use strict";
class smlValid extends HTMLElement {
  constructor() {
        super();
        let slv = this;
  }
  static observedAttributes=["data-api"];

    // connect component
  async connectedCallback() {
    let slv = this;
    slv.id = slv.id || "slv"+clip(guid(true),20);
    if(!slv.title) slv.title="sml Valid element not explained, ask DASM to fix this";
    if(!slv.ariaLabel) slv.ariaLabel=slv.title;
    if(slv.className=="") slv.className="smlValid text-danger fw-bold d-none m-2 p-2";
    slv.dataset.api = slv.getAttribute("data-api") || "";
    if(slv.dataset.api && !slv.dataset.apiForward) {
      slv.dataset.apiForward = slv.dataset.api.replaceAll("/","_").replaceAll("?","_").replaceAll("&","_").replaceAll("=","_");
      console.log("smlInput: Setting apiForward to "+slv.dataset.apiForward);
    }
    slv.dataset.icon= slv.getAttribute("data-icon") || "bi bi-question-circle";
    if(!slv.for) {
      let firstInput=slv.parentElement.querySelector("input[data-cc-property]");
      slv.for = firstInput ? firstInput.id : firstInput.name;
    }
    //outer span
    if(!slv.classList.contains("smlLbl")) slv.classList.add("smlLbl");
    //if(!slv.classList.contains("text-truncate")) slv.classList.add("text-truncate");
    if(!slv.classList.contains("text-break")) slv.classList.add("text-break");
    slv.wire();
  }

  // <div id="ReadLogLogAppFormGroup" class="form-floating border border-1 border-dark m-2 p-2" style="min-width:24em;min-height:.5em;flex-direction: column;justify-content:center;">
  //   <label for="ReadLogLogAppFormField" id="ReadLogLogAppLabel" aria-hidden="true" class="mt-0" style="padding: 0;" title="Access to the Log App" aria-label="Access to the Log App" t="Log App&nbsp;&nbsp;">
  //     Log App&nbsp;&nbsp;
  //   </label>
  //   <div id="ReadLogLogAppEnvelope" class="col row">
  //   <div type="text" id="ReadLogLogAppFormField" class="mt-4 px-2 fw-bold form-control-sm pt-3" data-cc-property="LogApp" style="max-width:21.2em;" title="Access to the Log App" aria-label="Access to the Log App" placeholder="Log App" readonly="true" t="DotnetCats">
  //      DotnetCats
  //   </div>
  //  </div>
  //  <div id="ReadLogLogAppFormFieldValid" class="d-none text-danger" t="">
  //  </div>
  // </div>



  // attribute change handler
  /**
   * This updates class properties based on html tags being altered
   *
   * @param {*} name name of element tag
   * @param {*} oldValue old value for reference of element tag
   * @param {*} newValue new value of element tag
   * @memberof ccSPA
   */
  async attributeChangedCallback(name, oldValue, newValue) {
      let cctg=this;
      switch(name){
              case "data-api":
                  cctg.api=newValue;
                  break;
              default: //WATDUH!
                  console.log(`smlInput: What you talking bout Willis? for attribute change ${name} from ${oldValue} to ${newValue}`);
                  break;
      }
  }

  wire() {
      let slv = this;
      //For API Calls
      if(slv.dataset.api!=="") {
          slv.addEventListener("click", async (e) => { slv.apiCall(e);});
      }
      //For Url (Navigation)
      if(slv.dataset.url) {
          slv.addEventListener("click", async (e) => {
              unobtrusiveWait("Going to " + e.currentTarget.dataset.url);
              window.location.href=slv.dataset.url;
              setTimeout(()=>{unobtrusiveWaitOff();},10000);
          });
      }
  }

  async apiCall() {
    let slv = this;
    //api call logic here
    unobtrusiveWait("Processing " + slv.dataset.text);
    let postData=await apiPostDirect(slv.dataset.api, { forward: slv.dataset.apiForward });
    unobtrusiveWaitOff();
    if (postData && postData.receipt && receiptCheckGood(postData.receipt)) {
        //success logic here
        window[slv.dataset.apiForward](postData);
    } else {
        //failure logic here
        console.log("smlInput: API Call failed for "+slv.dataset.api);
    }
    
  }




}

customElements.define("sml-valid", smlValid);

export default smlValid;

//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! S.D.G !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
//^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^