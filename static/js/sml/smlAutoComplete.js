//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! J.J. !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
//^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
/*!
 * smlAutoComplete --- Auto-complete input component for SML
 * Can be used standalone OR as a facade inside smlInput
 * Public Domain Licensed Copyright Law of the United States of America, Section 105 (https://www.copyright.gov/title17/92chap1.html#105)
 * Published by: Dominic Roche of OIT/IUSG/DASM
 * @class smlAutoComplete
 * @extends {HTMLElement}
 */
"use strict";
import { asFieldNotationString, isJson, isJsonRepaired } from "../smlUtils.js";
import smlFormField from "./smlFormField.js";
class smlAutoComplete extends HTMLElement {
  constructor() {
    super();
    const sac = this;
    sac.abortController = null;
    sac.minChars = 3;
    sac.debounceMs = 300;
    sac.isLoading = false;
    sac.isFacade = false; // Will be true if used inside smlInput
    sac.hasValidSelection = false;
    //ids
    sac.baseId = "";
    sac.searchEle = null;
    sac.selectEle = null;
    sac.hiddenEle = null;
  }

  connectedCallback() {
    const sac = this;
    sac.id = sac.id || "sac" + Math.random().toString(36).substring(2, 15);
    sac.baseId = sac.id;

    // Check if we're being used as a facade inside smlInput
    const parentInput = sac.closest('sml-input');
    
    sac.isFacade = !!parentInput;
    let property="";
    sac.id= parentInput._inputId;// + "AC";
    sac.searchEle = sac.id + "TextSearch";
    sac.selectEle = sac.id + "Facade";
    sac.hiddenEle = sac.id + "Hidden";
    if(sac.isFacade) {
      sac.property = parentInput.dataset.ccProperty;
      sac.dataset.smlAcType="SFF";
    } else {
      sac.property = sac.dataset.ccProperty;
      sac.dataset.smlAcType="SFFAC";
    }
    const api = sac.dataset.api;
    const label = sac.dataset.label || sac.property;
    
    // Allow override of min characters
    if (sac.dataset.minChars) {
      sac.minChars = parseInt(sac.dataset.minChars);
    }
    
    // Select type: 'single' (default) or 'multiple'
    sac.selectType = sac.dataset.selectType || 'single';
    
    if (!sac.property) {
      console.error('smlAutoComplete requires data-cc-property');
      return;
    }

    if (!api) {
      console.error('smlAutoComplete requires data-api');
      return;
    }

    sac.render(sac.property, label);
    
    sac.cacheElements();
    sac.attachEvents();
  }

  render(property, label) {
    const sac = this;
    const isMultiple = sac.selectType === 'multiple';
    const multipleAttr = isMultiple ? 'multiple' : '';
    const sizeAttr = isMultiple ? '6' : '3';
    // If used as facade, don't include label or outer wrapper
    if (sac.isFacade) {
      sac.innerHTML = `
        <input 
          type="search" 
          id="${sac.searchEle}"
          class="sml-ac-input smlActiveInput"
          placeholder="Type to search..."
          autocomplete="off"
        />
        
        <select 
          id="${sac.selectEle}"
          class="sml-ac-select hidden"
          size="${sizeAttr}"
          ${multipleAttr}
        >
          <option value="">~~~SELECT~~~</option>
        </select>
      `;
    } else {
      // Standalone mode: include everything
      sac.innerHTML = `
        <div class="smlFF d-flex flex-row shadow shadow-navy shadow-lg rounded-2 m-1 p-3 sfftext" style="flex-wrap: wrap;">
          <label for="${sac.searchEle}" style="top: -0.10rem;left: .5rem;font-size: .90rem;background: #fff;padding: 0 .15rem;color: #495057;z-index: 3;">${label}</label>
          <input 
            type="text" 
            id="${sac.searchEle}"
            class="sml-ac-input"
            placeholder="Type to search..."
            autocomplete="off"
          />
          
          <select 
            id="${sac.selectEle}" 
            class="sml-ac-select hidden"
            size="${sizeAttr}"
            ${multipleAttr}
          >
            <option value="">~~~SELECT~~~</option>
          </select>
          
          <input 
            type="hidden" 
            id="${sac.hiddenEle}"
            name="${sac.hiddenEle}"
            data-cc-property="${property}"
          />
        </div>
      `;
    }
  }

  cacheElements() {
    const sac = this;
    const prop = sac.dataset.property;

    sac.textInput = sac.querySelector(`#${sac.searchEle}`);
    sac.selectElement = sac.querySelector(`#${sac.selectEle}`);
    
    // If facade, the hidden input is in the parent smlInput
    if (sac.isFacade) {
      const parentInput = sac.closest('sml-input');
      // Prefer matching the real backing input by property, then fall back to any hidden
      sac.hiddenInput = parentInput?.querySelector(
      `input[type="hidden"][data-cc-property="${prop}"]`
      ) || parentInput?.querySelector('input[type="hidden"]');
    }


    const parentInput = sac.closest('sml-input');
    if (sac.isFacade) {
      sac.hiddenInput = parentInput?.querySelector("input[data-cc-property]");
    } else {
      sac.hiddenInput = sac.querySelector(`#${prop}`);
    }
  }

  attachEvents() {
    const sac = this;
    let debounceTimer;
    // Debounced search on input
    sac.textInput.addEventListener('input', (e) => {
      clearTimeout(debounceTimer);
      const value = e.target.value.trim();
      
      // If cleared, reset everything
      if (value === '') {
        sac.reset();
        return;
      }
      
      // Wait for minimum characters
      if (value.length < sac.minChars) {
        sac.hideSelect();
        return;
      }
      
      debounceTimer = setTimeout(() => sac.search(value), sac.debounceMs);
    });

    // Handle selection from select
    sac.selectElement.addEventListener('change', (e) => {
      if (sac.selectType === 'multiple') {
        sac.handleMultipleSelection(e);
      } else {
        sac.handleSingleSelection(e);
      }
    });

    // Hide select when both input and select lose focus
    sac.textInput.addEventListener('blur', () => {
      setTimeout(() => {
        if (document.activeElement !== sac.selectElement && 
            document.activeElement !== sac.textInput) {
          sac.hideSelect();
        }
      }, 150);
    });

    sac.selectElement.addEventListener('blur', () => {
      setTimeout(() => {
        if (document.activeElement !== sac.selectElement && 
            document.activeElement !== sac.textInput) {
          sac.hideSelect();
        }
      }, 150);
    });

    // Allow keyboard navigation: ArrowDown to move to select
    sac.textInput.addEventListener('keydown', (e) => {
      if (e.key === 'ArrowDown' && !sac.selectElement.classList.contains('hidden')) {
        e.preventDefault();
        sac.selectElement.focus();
        if (sac.selectElement.options.length > 1) {
          sac.selectElement.selectedIndex = 1;
        }
      }
    });

    // Allow keyboard navigation: ArrowUp to move back to input
    sac.selectElement.addEventListener('keydown', (e) => {
      if (e.key === 'ArrowUp' && sac.selectElement.selectedIndex === 0) {
        e.preventDefault();
        sac.textInput.focus();
      }
    });
  }

  async search(query) {
    const sac = this;
    // Cancel previous request
    sac.abortController?.abort();
    sac.abortController = new AbortController();

    // Show loading state
    sac.showLoading();

    try {
      const data = await sac.fetchResults(query, sac.abortController.signal);
      sac.populateSelect(data);
      sac.showSelect();
    } catch (err) {
      if (err.name === 'AbortError') return;
      console.error('Search failed:', err);
      sac.showError(err.message);
    } finally {
      sac.isLoading = false;
    }
  }

  async fetchResults(query, signal) {
    const sac = this;
    const apiUrl = sac.dataset.api;
    const property = sac.dataset.ccProperty;
    
    const payload = {
      searchString: query,
      codeCaller: 'smlAutoComplete',
      callerProperty: property
    };

    // SEND UP: Add filter values from other form fields
    if (sac.dataset.apiFiltersUp) {
      const filters = sac.dataset.apiFiltersUp.split(',').map(f => f.trim());
      filters.forEach(filter => {
        const filterEl = document.querySelector(`[data-cc-property="${filter}"]`);
        if (filterEl) {
          payload[filter] = filterEl.value;
        }
      });
    }

    // SEND UP: Tell API which properties we want back
    if (sac.dataset.apiPropsDown) {
      const [isValid, repaired] = isJsonRepaired(sac.dataset.apiPropsDown);
      if (isValid) {
        const propsDown = JSON.parse(repaired);
        payload.requestedProperties = Object.values(propsDown);
      } else {
        //Assuming comma delimited list of properties
        payload.requestedProperties = sac.dataset.apiPropsDown;
      }


    }
    const dataUp=JSON.stringify(payload);
    const response = await fetch(apiUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: dataUp,
      signal
    });

    if (!response.ok) {
      throw new Error(`API error: ${response.status} ${response.statusText}`);
    }
    
    sac.data = await response.json();
    
    if (sac.data.errorObject) {
      throw new Error(data.errorObject);
    }
    
    return sac.data;
  }

  populateSelect(data) {
    const sac = this;
    // Always start with ~~~SELECT~~~ option
    sac.selectElement.innerHTML = '<option value="">~~~SELECT~~~</option>';
    
    if (!data || data.length === 0) {
      const noResults = document.createElement('option');
      noResults.textContent = 'No results found';
      noResults.disabled = true;
      sac.selectElement.appendChild(noResults);
      return;
    }

    const apiValue = sac.dataset.apiValue || 'text';
    const apiId = sac.dataset.apiId || 'id';

    data.forEach(item => {
      const option = document.createElement('option');
      option.value = apiId ? item[apiId] : item[apiValue];
      option.textContent = item[apiValue];
      
      if (apiId && item[apiId]) {
        option.dataset.id = item[apiId];
      }
      
      // SEND DOWN PREP: Store properties as data attributes
      if (sac.dataset.apiPropsDown) {
        const [isValid, repaired] = isJsonRepaired(sac.dataset.apiPropsDown);
        if(!isValid) {
          const propsDown = sac.dataset.apiPropsDown;

        }else { //Probably not going to use this...list of property names is simpler
          const propsDown = JSON.parse(repaired);
          Object.entries(propsDown).forEach(([formField, apiProperty]) => {
            if (item[apiProperty] !== undefined && item[apiProperty] !== null) {
              option.dataset[sac.camelCase(apiProperty)] = item[apiProperty];
            }
          });
        }
      }
      sac.selectElement.appendChild(option);
    });
  }

  showLoading() {
    const sac = this;
    sac.isLoading = true;
    sac.selectElement.innerHTML = `
      <option value="">~~~SELECT~~~</option>
      <option disabled>⏳ Loading results...</option>
    `;
    sac.showSelect();
  }

  showError(message) {
    const sac = this;
    sac.selectElement.innerHTML = `
      <option value="">~~~SELECT~~~</option>
      <option disabled>❌ Error: ${message}</option>
    `;
    sac.showSelect();
  }

  showSelect() {
    const sac = this;
    sac.selectElement.classList.remove('hidden');
  }

  hideSelect() {
    const sac = this;
    sac.selectElement.classList.add('hidden');
    
    // Trigger validation if text exists but no valid selection
    if (!sac.hasValidSelection && sac.textInput.value.trim() !== '') {
      if (sac.hiddenInput) {
        sac.hiddenInput.value = ''; // Clear to make it invalid
        sac.hiddenInput.dispatchEvent(new Event('blur', { bubbles: true }));
      }
    }
  }

  reset() {
    const sac = this;
    sac.hasValidSelection = false;
    if (sac.hiddenInput) {
      sac.hiddenInput.value = '';
    }
    sac.hideSelect();
    sac.selectElement.innerHTML = '<option value="">~~~SELECT~~~</option>';
  }


  clearSelection() {
    const sac = this;
    if (sac.hiddenInput) {
      sac.hiddenInput.value = '';
    }
    sac.textInput.value = '';
    sac.textInput.focus();
    
    // Clear related fields
    if (sac.dataset.apiPropsDown) {
      const propsDown = JSON.parse(sac.dataset.apiPropsDown);
      Object.keys(propsDown).forEach(formField => {
        const targetField = document.querySelector(`[data-cc-property="${formField}"]`);
        if (targetField) {
          targetField.value = '';
        }
      });
    }
  }

  populateRelatedFields(selectedOption) {
    const sac = this;
    if (!sac.dataset.apiPropsDown) return;
    

    const [isValid, repaired] = isJsonRepaired(sac.dataset.apiPropsDown);
    if(!isValid) {
      const propsDown = sac.dataset.apiPropsDown;
      propsDown.split(',').forEach(prop => {
        if (prop !== undefined) {
          const targetField = document.querySelector(`input[data-cc-property="${prop}"]`);
          
          if (targetField) {
            //targetField.value = sac.data[sac.querySelector("select").selectedIndex-1][prop] || '';
            targetField.value = sac.data[sac.querySelector("select").selectedIndex-1][asFieldNotationString(prop)]||'';
            targetField.dispatchEvent(new Event('change', { bubbles: true }));
          }
        }
      });
    }else{
      const propsDown = JSON.parse(repaired);
      //Object.entries(propsDown).forEach(([formField, apiProperty]) => {
        for(let [formField, apiProperty] of Object.entries(propsDown)) {
        const dataKey = sac.camelCase(apiProperty);
        const value = selectedOption.dataset[dataKey];
        if (value !== undefined) {
          const targetField = document.querySelector(`[data-cc-property="${formField}"]`);
          
          if (targetField) {
            targetField._input.value = value;
            targetField._input.dispatchEvent(new Event('change', { bubbles: true }));
          }
        }
      };
    }
  }

  camelCase(str) {
    const sac = this;
    return str.replace(/[-_](.)/g, (_, char) => char.toUpperCase());
  }

handleSingleSelection(e) {
  const sac = this;
  const selectedOption = e.target.options[e.target.selectedIndex];
 
  if (selectedOption.value === '') {
    sac.clearSelection();
    sac.hasValidSelection = false;
    return;
  }
  
  sac.hasValidSelection = true;
  
  // Update visible text input
  sac.textInput.value = selectedOption.textContent;
  
  // Update hidden input
  if (sac.dataset.smlAcType === "SFFAC") {
      sac.hiddenInput = sac?.querySelector("input[data-cc-property]");
      if(!sac.hiddenInput) {
        const parentInput = sac.closest('sml-input');
        sac.hiddenInput = parentInput?.querySelector("input[data-cc-property]");
      }
  }else{
      sac.hiddenInput = sac?.querySelector("input[data-cc-property]");
      if(!sac.hiddenInput) {
        const parentInput = sac.closest('sml-input');
        sac.hiddenInput = parentInput?.querySelector("input[data-cc-property]");
      }
  }
  if (sac.hiddenInput) {
    sac.hiddenInput.value = selectedOption.value;
    
    if (selectedOption.dataset.id) {
      sac.hiddenInput.dataset.selectedId = selectedOption.dataset.id;
    }
    
    // Trigger smlInput's change handler to add smlHasValue class
    sac.hiddenInput.dispatchEvent(new Event('change', { bubbles: true }));
  }
  
  if(sac.selectType==="single") sac.hideSelect();

  // SEND DOWN: Populate other fields
  sac.populateRelatedFields(selectedOption);
  
  // Dispatch event
  sac.dispatchEvent(new CustomEvent('sml-selected', {
    detail: {
      value: selectedOption.value,
      text: selectedOption.textContent,
      data: selectedOption.dataset
    },
    bubbles: true
    }));
  }

  handleMultipleSelection(e) {
    const sac = this;
    const selectedOptions = Array.from(e.target.selectedOptions)
      .filter(opt => opt.value !== '');
    
    if (selectedOptions.length === 0) {
      sac.clearSelection();
      return;
    }
    
    const texts = selectedOptions.map(opt => opt.textContent).join(', ');
    sac.textInput.value = texts;
    
    const values = selectedOptions.map(opt => opt.value).join(',');
    if (sac.hiddenInput) {
      sac.hiddenInput.value = values;
      
      const ids = selectedOptions
        .map(opt => opt.dataset.id)
        .filter(id => id)
        .join(',');
      if (ids) {
        sac.hiddenInput.dataset.selectedId = ids;
      }
      
      sac.hiddenInput.dispatchEvent(new Event('change', { bubbles: true }));
    }
    
    sac.dispatchEvent(new CustomEvent('sml-selected', {
      detail: {
        value: values,
        text: texts,
        items: selectedOptions.map(opt => ({
          value: opt.value,
          text: opt.textContent,
          data: opt.dataset
        }))
      },
      bubbles: true
    }));
  }

  // Public API
  setValue(value, text) {
    const sac = this;
    if (sac.hiddenInput) {
      sac.hiddenInput.value = value;
    }
    sac.textInput.value = text;
    sac.hideSelect();
  }

  getValue() {
    const sac = this;
    return {
      value: sac.hiddenInput?.value || '',
      text: sac.textInput.value
    };
  }
}

customElements.define('sml-auto-complete', smlAutoComplete);

export default smlAutoComplete;
//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! S.D.G !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
//^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^