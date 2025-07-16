Prince.trackBoxes = true;
Prince.registerPostLayoutFunc(insertPageBreaks(1));

const headerTemplate = document.getElementById("mp-header-templ");
const footerTemplate = document.getElementById("mp-footer-templ");

function insertPageBreaks(fromPage) {
  return function () {
    const tail = document.querySelector(".tail");
    const boxes = tail.getPrinceBoxes();
    const tailPage = boxes[0].pageNum;

    var inserted = false;
    for (let page = fromPage; page < tailPage; page++) {
      insertTemplate(footerTemplate, page);
      insertTemplate(headerTemplate, page + 1);
      inserted = true;
    }

    if (fromPage == 1) {
      // We've just inserted page breaks for the first time, we need to do
      // another pass, after Prince has refreshed the layout, just in case the
      // tail was pushed further, which would require more page breaks.
      Prince.registerPostLayoutFunc(insertPageBreaks(tailPage));
    } else {
      // We've made two passes of inserting page breaks, we continue by updating
      // the amounts in the page breaks.
      if (inserted) {
        // DOM has changed, we need to wait for Prince to refresh the layout
        // before we can update the amounts
        Prince.registerPostLayoutFunc(updatePageBreakAmounts);
      } else {
        // DOM has not changed, we must update the amounts immediately
        updatePageBreakAmounts();
      }
    }
  }
}

function updatePageBreakAmounts() {
  const lines = document.querySelectorAll("tr[data-subtotal]");
  if (lines.length == 0) return;

  let prevSubtotal = "";
  let prevPage = 1;

  // Iterate over the lines and update page break amounts
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const boxes = line.getPrinceBoxes();
    const page = boxes[0].pageNum

    if (page > prevPage) {
      updatePageBreakAmount("mp-footer-" + prevPage, prevSubtotal);
      updatePageBreakAmount("mp-header-" + page, prevSubtotal);
    }

    prevSubtotal = line.getAttribute("data-subtotal");
    prevPage = page;
  }

  // Iterate over the remaining pages until the tail and update the page breaks
  const totals = document.querySelector(".tail");
  const boxes = totals.getPrinceBoxes();
  const tailPage = boxes[0].pageNum;

  for (let page = prevPage; page < tailPage; page++) {
    updatePageBreakAmount("mp-footer-" + page, prevSubtotal);
    updatePageBreakAmount("mp-header-" + (page + 1), prevSubtotal);
  }
}

function insertTemplate(template, page) {
  const id = template.id.replace("-templ", "-" + page);
  const el = template.cloneNode(true);
  el.id = id;
  el.setAttribute("style", "-prince-float-defer-page: " + (page - 1));
  el.style.display = "block";

  const body = document.getElementsByTagName('body')[0];
  body.insertBefore(el, body.firstChild);
}

function updatePageBreakAmount(id, amount) {
  const el = document.getElementById(id);
  if (el) {
    el.querySelector(".amount").innerHTML = amount;
  }
}
