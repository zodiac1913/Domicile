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
 * @class smlAppButton
 * @extends {HTMLElement}
 */
// תהילתו. לא שלי
import sml from './sml.js';
import { clip, guid, jmlToHtml, unobtrusiveWait, unobtrusiveWaitOff } from "../smlUtils.js";
"use strict";
class smlAppButton extends HTMLElement {

    /* -~~--~~--~~--~~--~~--~~--~~--~~-- Init Lifecycle -~~--~~--~~--~~--~~--~~--~~- */ 

    static observedAttributes=["data-api","data-toggler-done","data-active"];

    async connectedCallback() {
        let sab = this;
        sab.initializeBaseAttributes();
        
    }



    /* -~~--~~--~~--~~--~~--~~--~~--~~-- END Init Lifecycle -~~--~~--~~--~~--~~--~~--~~- */ 


    /* -~~--~~--~~--~~--~~--~~--~~--~~-- Setup Calls -~~--~~--~~--~~--~~--~~--~~- */ 

    initializeBaseAttributes() {
        let sab = this;
        sab.id = sab.id || "sab"+clip(guid(true),20);
        if(!sab.type) sab.type="button";
        if(!sab.role) sab.role="button";
        if(!sab.hasAttribute("tabindex")) sab.tabIndex = 0;
        if(!sab.title) sab.title="sml App Button not explained, ask DASM to fix this";
        if(!sab.ariaLabel) sab.ariaLabel=sab.title;
        if(sab.className=="") sab.className="smlRB btn btn-outline-primary border-0 text-white text-nowrap text-truncate";
    }

    /* -~~--~~--~~--~~--~~--~~--~~--~~-- End Setup Calls -~~--~~--~~--~~--~~--~~--~~- */ 


}
