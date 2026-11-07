// ==UserScript==
// @name         Shoka Download Button
// @namespace    http://tampermonkey.net/
// @version      1.0
// @description  Adds button to nhentai archives to download with shoka.
// @author       Jqnx
// @match        *://nhentai.net/*
// @grant        none
// ==/UserScript==

(function () {
  "use strict";

  const btn = document.createElement("button");

  btn.textContent = "Add to Shoka Queue";
  btn.id = "shoka";
  btn.style.padding = "0px 12px";
  btn.style.height = "40px";
  btn.style.margin = "3px 5px 10px 0";
  btn.style.minWidth = "120px";
  btn.style.backgroundColor = "#6d28d9";
  btn.style.color = "white";
  btn.style.border = "0";
  btn.style.borderRadius = "3px";
  btn.style.cursor = "pointer";
  btn.style.textAlign = "center";
  btn.style.fontWeight = "700";
  btn.style.verticalAlign = "middle";

  btn.addEventListener("click", function () {
    alert("Download with shoka!");
  });

  const buttons = document.querySelector(".buttons");

  if (buttons) {
    buttons.appendChild(btn);
  }
})();

