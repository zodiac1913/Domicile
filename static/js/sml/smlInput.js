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
 * smlInput --- sml FormField Input module for SML FormFields
 * Public Domain Licensed Copyright Law of the United States of America, Section 105 (https://www.copyright.gov/title17/92chap1.html#105)
 * Et qui me misit, mecum est: non reliquit me solum Pater, quia ego semper quae placita sunt ei, facio!
 * Published by: Dominic Roche of OIT/IUSG/DASM on 1/5/2026
 * @class smlInput
 * @extends {HTMLElement}
 */
// תהילתו. לא שלי
import sml from '../sml/sml.js';
import { asBool, clip, guid, jmlToHtml, replaceLast, unobtrusiveWait, unobtrusiveWaitOff } from "../smlUtils.js";
"use strict";
class smlInput extends HTMLElement {
  constructor(fld) {
    super();
    let slInp = this;
    if(fld!==undefined){
      slInp.fld=fld;
    }
    this._input = null;

    for (const prop of this.constructor.forwardedProperties) {
      Object.defineProperty(this, prop, {
        get() {
          return this._input ? this._input[prop] : this.getAttribute(prop);
        },
        set(v) {
            this.setAttribute(prop, v);
        },
      });
    }
  }
  static observedAttributes=["data-api"];

    // connect component
  async connectedCallback() {
    let slInp = this;
    const genId=clip(guid(true),10);
    slInp.id = (slInp.id || "slInp"+genId+"SmlInput");
    slInp._inputId=replaceLast(replaceLast(replaceLast(slInp.id,"SmlFFSmlInput",""),"Input",""),"Sml","");
    if(!slInp.id.endsWith("SmlInput")) slInp.id=slInp.id+"SmlInput";
    slInp.name=slInp.id;
    
    //get form action type
    let readOnly=false;
    let formEle=document.querySelector(".ccFormCard");
    let formType=formEle?.dataset?.action;
    if(formType=="Read" || formType=="Delete") readOnly=true;

    // Create the real input 
    //slInp._inputId = slInp.id + (slInp.id.includes("FormField")?"":"Input"); 
    slInp.innerHTML = `<input id="${slInp._inputId}">`;
    slInp.io="input";
    slInp._input = slInp.querySelector(slInp.io);
    slInp._input.id = slInp._inputId;
    slInp._input.type = "text";//slInp.dataset.type ||"text";
    if(slInp.dataset.type.toLowerCase()==="hidden") {
      slInp.dataset.hidden="true";
      slInp._input.className="";
      slInp._input.style.display="none";
    }else{
      slInp._input.className = "smlInpt smlFloat flex-fill mt-4 px-2 fw-bold";
    }
    slInp._input.placeholder = " ";
    slInp._input.dataset.ccProperty=slInp.dataset.ccProperty || "Unknown";
    slInp._input.isRequired=asBool(slInp.dataset.isRequired)||false;
    if (!slInp.hasAttribute("title")) {
      this._input.setAttribute("title", "Form field");
      console.warn("sml Input missing title — DASM, fix your stuff.");
    }
    if (!this._input.hasAttribute("aria-label") &&
        !this._input.hasAttribute("aria-labelledby")) {
      this._input.setAttribute("aria-label", this._input.title);
    }
    if(slInp.className=="") slInp.className="smlInpt flex-fill";

    slInp.dataset.api = slInp.getAttribute("data-api") || "";
    if(slInp.dataset.api && !slInp.dataset.apiForward) {
      slInp.dataset.apiForward = slInp.dataset.api.replaceAll("/","_").replaceAll("?","_").replaceAll("&","_").replaceAll("=","_");
      console.log("smlInput: Setting apiForward to "+slInp.dataset.apiForward);
    }
    slInp.dataset.icon= slInp.getAttribute("data-icon") || "bi bi-question-circle";
    if (slInp.dataset.text) {
      slInp._input.value = slInp.dataset.text;
    }
    slInp.dataset.text = "";
    //Find your label
    let firstLabel = this.parentElement?.querySelector("sml-label,label");
    if (firstLabel) {
      // Ensure label has an ID
      if (!firstLabel.id) firstLabel.id = this.id + "Label";
      // Wire label to the real input
      firstLabel.setAttribute("for", slInp._input.id);
      slInp._input.setAttribute("aria-labelledby", firstLabel.id);
    }
    if(!slInp.classList.contains("smlInput")) slInp.classList.add("smlInput");
    // Forward initial attributes 
    slInp._forwardInitialAttributes();
    let fld={};
    fld.name=slInp._input.dataset.ccProperty || "Unknown";
    fld.title=slInp._input.title || "Form field";
    fld.defaultValue=slInp._input.value || "";
    fld.sfDefaultValueShow=slInp.dataset.sfDefaultValueShow || "";
    fld.isRequired=slInp._input.required || false;
    fld.label=firstLabel?.dataset?.text || firstLabel?.textContent || "Some field I didnt bother to label?";
    fld.icon=slInp.dataset.icon || "bi bi-question-circle";
    if(!slInp.type) { slInp.dataset.type="text"; }
    //)))))))))))))))))))))))))))))))))))))Add Facades for Input when needed(((((((((((((((((((((((((((((
    let inputFacade={};
    switch(slInp.dataset.type.toLowerCase()){ 
      case "smlautocomplete":
          // Create sml-auto-complete as facade
          const smlAc = document.createElement('sml-auto-complete');
          smlAc.id = slInp._inputId + "AC";
          smlAc.name = smlAc.id;
          smlAc.className = "d-flex flex-column smlInput smlActiveInput smlFloat flex-fill mt-4 px-2 fw-bold sffautocomplete";
          smlAc.dataset.ccProperty = slInp.dataset.ccProperty;
          smlAc.dataset.label = slInp.parentElement.dataset.label;
          smlAc.dataset.api = slInp.dataset.api || slInp.parentElement.dataset.api;
          smlAc.dataset.apiValue = slInp.dataset.apiValue || slInp.parentElement.dataset.apiValue;
          smlAc.dataset.apiId = slInp.dataset.apiId || slInp.parentElement.dataset.apiId;
          smlAc.dataset.url = slInp.dataset.url || slInp.parentElement.dataset.url || "";
          smlAc.dataset.apiPropsDown = slInp.dataset.apiPropsDown || slInp.parentElement.dataset.apiPropsDown || "";
          smlAc.dataset.apiFiltersUp = slInp.dataset.apiFiltersUp || slInp.parentElement.dataset.apiFiltersUp || "";
          smlAc.inputFacade=[{n:"input",type: "text",i:slInp._inputId + "Facade",name: slInp._inputId + "Facade"},
            {n:"select",i:slInp._inputId + "Select",name: slInp._inputId + "Select",size:5}
          ];

          slInp.parentElement.parentElement.querySelector("sml-label").style.left = ".5rem";
          slInp.parentElement.parentElement.querySelector("sml-label").style.top="-0.10rem"
          slInp.parentElement.parentElement.querySelector("sml-label").style.fontSize = ".90rem";
          slInp.parentElement.parentElement.querySelector("sml-label").style.background = "#fff";
          slInp.parentElement.parentElement.querySelector("sml-label").style.padding = "0 .15rem";
          slInp.parentElement.parentElement.querySelector("sml-label").style.color = "#495057";
          slInp.parentElement.parentElement.querySelector("sml-label").style.zIndex = "3";

          
          // Optional attributes
          if (slInp.dataset.apiPropsDown || slInp.parentElement.dataset.apiPropsDown) {
              smlAc.dataset.apiPropsDown = slInp.dataset.apiPropsDown || slInp.parentElement.dataset.apiPropsDown;
          }
          if (slInp.dataset.apiFiltersUp || slInp.parentElement.dataset.apiFiltersUp) {
              smlAc.dataset.apiFiltersUp = slInp.dataset.apiFiltersUp || slInp.parentElement.dataset.apiFiltersUp;
          }
          if (slInp.dataset.ccSelectType || slInp.parentElement.dataset.ccSelectType) {
              smlAc.dataset.ccSelectType = slInp.dataset.ccSelectType || slInp.parentElement.dataset.ccSelectType;
          }
          slInp.dataset.hasFacade = "true";
          slInp.insertAdjacentElement('afterBegin', smlAc);
          slInp._input.type = "hidden";
          break;      
      case "select":
          inputFacade={n:"select",type: slInp.dataset.type,i:slInp._inputId + "Facade",name: slInp._inputId + "Facade"
            ,c:"smlInput smlActiveInput align-items-left flex-fill form-select fw-bold mt-4 sffselect",b:[]};
          if(slInp.dataset.sfDropDownJson){
              let optionsData=JSON.parse(slInp.dataset.sfDropDownJson);
              for(let opt of optionsData){
                  inputFacade.b.push({n:"option",value:opt.id,t:opt.text||opt.value});
              }
          }
          slInp._input.type="hidden";
          slInp.dataset.hasFacade="true";
        break;
      case "textarea":
          inputFacade={n:"textarea",type: slInp.dataset.type + "Facade",i:slInp._inputId+"Facade",name: slInp._inputId + "Facade",c:"smlActiveInput fw-bold align-items-left smlInpt flex-fill",placeholder: " ",value:slInp.value};
          slInp._input.type="hidden";
          slInp.dataset.hasFacade="true";
        break;
      case "checkbox":
          inputFacade={n:"input",i:slInp._inputId + "Facade",name: slInp._inputId + "Facade"
              ,c:"smlInput smlActiveInput form-check-input mt-4 px-2 fw-bold ms-2 sffcheckbox px-2 fw-bold ms-2" //smlFloat 
              ,"data-label":fld.label,"data-icon":fld.icon,"data-text":fld.defaultValue
              ,placeholder: " "
              ,"data-property":fld.name,style:"width:1em !Important;padding: .50rem .50rem .50rem;appearance: checkbox;margin-left: 1.5rem !important;"
              ,"type":"checkbox"
              ,"data-is-required":fld.isRequired?"true":"false"
              ,value:asBool(fld.defaultValue)?"true":"false"
              ,title: fld.title
          };
          slInp._input.type="hidden";
          slInp.dataset.hasFacade="true";
        break;
      default: //text, password, email, number, date, etc
            // inputFacade={n:"input",type: slInp.dataset.type,role:slInp.dataset.type
            //       ,i:slInp._inputId + "Facade",c:"smlInput smlFloat smlActiveInput fw-bold align-items-left flex-fill mt-4"
            //       ,"data-intype": slInp.dataset.type,value: fld.defaultValue
            //       ,value:slInp.value};
            //^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^Avoid Facades for normal inputs^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^            
            slInp.dataset.hasFacade="false";
        break;            
    }
    const inputType = (slInp.dataset.type || "text").toLowerCase();
    if(readOnly && inputFacade?.n) inputFacade.readonly="true";

    if(inputFacade?.n){
      const inpt=jmlToHtml(inputFacade);
      if(inpt) slInp.insertAdjacentHTML("afterBegin", inpt);
    }

    if(inputType==="text"){
      slInp._input.classList.add("smlActiveInput");
    }
    if(slInp.dataset.defaultValue && slInp.dataset.defaultValue!="") {
      const facadeInput=slInp.querySelector("#"+slInp._inputId + "Facade");
      if(facadeInput){
        facadeInput.value=slInp.dataset.defaultValue;
      }else{
        slInp._input.value=slInp.dataset.defaultValue;
      }
    }
    //)))))))))))))))))))))))))))))))))))))END Facades for Input when needed(((((((((((((((((((((((((((((
    slInp.wire();
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
static get forwardedProperties() {
  return [
    "value",
    "placeholder",
    "required",
    "disabled",
    "readOnly",
    "maxLength",
    "minLength",
    "pattern",
    "autocomplete",
    "name",
    "type",
    "checked",
    "selectedIndex"
  ];
}


  _forwardInitialAttributes() {
    const slInp = this;

    // Attributes we do NOT want to forward
    const blockList = new Set([
      "id",
      "class",
      "style",
      "data-api",
      "data-api-forward",
      "data-icon",
      "data-text",
      "data-url",
      "type" // handled separately
    ]);

    for (const attr of slInp.getAttributeNames()) {
      if (!blockList.has(attr)) {
        const value = slInp.getAttribute(attr);
        slInp._input.setAttribute(attr, value);
      }
    }

    // Special case: forward type to the inner input
    if (slInp.type) {
        slInp._input.setAttribute("type", slInp.type);
      }
    }



  wire() {
      let slInp = this;
      //For API Calls
      if(slInp.dataset.api!=="") {
          slInp.addEventListener("click", async (e) => { slInp.apiCall(e);});
      }

      //Fill in facade if input directly altered
      slInp._input.addEventListener("change", async (e) => {
          slInp.dataset.text = e.target.value;
          let facadeEle=slInp.parentNode.querySelector("#"+slInp._inputId+"Facade");

          if(facadeEle){
              if(slInp.dataset.type.toLowerCase()=="checkbox"){
                  facadeEle.checked=asBool(e.target.value);
              }else{
                  facadeEle.value=e.target.value;
              }
          }else{
            if (slInp._input.value && slInp._input.value.trim() !== "" && slInp._input.type.toLowerCase()!="hidden") { 
              slInp._input.classList.add("smlHasValue"); 
            } else { 
              slInp._input.classList.remove("smlHasValue"); 
            }

          }
          if(facadeEle===null){ //facade not found
            if (slInp._input.value && slInp._input.value.trim() !== "") { 
              slInp._input.classList.add("smlHasValue"); 
            } else { 
              slInp._input.classList.remove("smlHasValue"); 
            }
          }else{
            if (facadeEle.value && facadeEle.value.trim() !== "") { 
              facadeEle.classList.add("smlHasValue"); 
            } else { 
              facadeEle.classList.remove("smlHasValue"); 
            }
         }
      });


      //For Url (Navigation)
      if(slInp.dataset.url) {
          slInp.addEventListener("click", async (e) => {
              unobtrusiveWait("Going to " + e.currentTarget.dataset.url);
              window.location.href=slInp.dataset.url;
              setTimeout(()=>{unobtrusiveWaitOff();},10000);
          });
      }else if(slInp.dataset.type.toLowerCase()=="select") {  //SELECT
          slInp.parentNode.querySelector("select").addEventListener("change", async (e) => {
              slInp._input.value = e.target.value;
          });
      }else if(slInp.dataset.type.toLowerCase()=="checkbox") {//Checkbox
          slInp.parentNode.querySelector("input[type=checkbox]").addEventListener("change", async (e) => {
              slInp._input.value = asBool(e.target.checked);
          });
      }else{
          slInp.parentNode.querySelector("input").addEventListener("change", async (e) => {
          let inputEle=e.currentTarget;
          //inputEle.type="text";
          slInp._input.value = inputEle.value;
          if(inputEle.value && inputEle.value.trim()!=="") { inputEle.classList.add("smlHasValue"); } else { inputEle.classList.remove("smlHasValue"); }
        });
      }
      //Set Default Value if exists
      //--------------------------Input      
      if(slInp.dataset.defaultValue && slInp.dataset.defaultValue!="") { 
        slInp._input.value=slInp.dataset.defaultValue;
      }
      //--------------------------Facade
      let facadeEle=slInp.parentNode.querySelector("#"+slInp._inputId+"Facade");
      if(facadeEle){
          if(slInp.dataset.type.toLowerCase()=="checkbox"){
              facadeEle.checked=asBool(slInp.dataset.defaultValue);
          }else{
              facadeEle.value=slInp.dataset.defaultValue;
          }
      }

  }

  async apiCall() {
    let slInp = this;
    //api call logic here
    unobtrusiveWait("Processing " + slInp.dataset.text);
    let postData=await apiPostDirect(slInp.dataset.api, { forward: slInp.dataset.apiForward });
    unobtrusiveWaitOff();
    if (postData && postData.receipt && receiptCheckGood(postData.receipt)) {
        //success logic here
        window[slInp.dataset.apiForward](postData);
    } else {
        //failure logic here
        console.log("smlInput: API Call failed for "+slInp.dataset.api);
    }
    
  }

  // --- Core value API ---
  get value() {
    return this._input?.value ?? "";
  }
  set value(v) {
    if (this._input) this._input.value = v;
  }


  // --- Placeholder ---
  get placeholder() {
    return this._input?.placeholder ?? "";
  }
  set placeholder(v) {
    if (this._input) this._input.placeholder = v;
  }

  // --- Required ---
  get required() {
    return this._input?.required ?? false;
  }
  set required(v) {
    if (this._input) this._input.required = v;
  }

  // --- Disabled ---
  get disabled() {
    return this._input?.disabled ?? false;
  }
  set disabled(v) {
    if (this._input) this._input.disabled = v;
  }

  // --- Readonly ---
  get readOnly() {
    return this._input?.readOnly ?? false;
  }
  set readOnly(v) {
    if (this._input) this._input.readOnly = v;
  }

  // --- Maxlength ---
  get maxLength() {
    return this._input?.maxLength ?? -1;
  }
  set maxLength(v) {
    if (this._input) this._input.maxLength = v;
  }

  // --- Minlength ---
  get minLength() {
    return this._input?.minLength ?? -1;
  }
  set minLength(v) {
    if (this._input) this._input.minLength = v;
  }

  // --- Pattern ---
  get pattern() {
    return this._input?.pattern ?? "";
  }
  set pattern(v) {
    if (this._input) this._input.pattern = v;
  }

  // --- Autocomplete ---
  get autocomplete() {
    return this._input?.autocomplete ?? "";
  }
  set autocomplete(v) {
    if (this._input) this._input.autocomplete = v;
  }

  // --- Name ---
  get name() {
    return this._input?.name ?? "";
  }
  set name(v) {
    if (this._input) this._input.name = v;
  }

  // --- Type (for input elements only) ---
  get type() {
    return this._input?.type ?? "";
  }
  set type(v) {
    if (this._input && "type" in this._input) {
      this._input.type = v;
    }
  }

  // --- Checked (checkbox / radio only) ---
  get checked() {
    return this._input?.checked ?? false;
  }
  set checked(v) {
    if (this._input && "checked" in this._input) {
      this._input.checked = v;
    }
  }

  // --- SelectedIndex (select only) ---
  get selectedIndex() {
    return this._input?.selectedIndex ?? -1;
  }
  set selectedIndex(v) {
    if (this._input && "selectedIndex" in this._input) {
      this._input.selectedIndex = v;
    }
  }

  // --- Options (select only) ---
  get options() {
    return this._input?.options ?? null;
  }



}

customElements.define("sml-input", smlInput);

export default smlInput;

//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! S.D.G !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
//^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^