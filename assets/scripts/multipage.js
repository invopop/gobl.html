Prince.trackBoxes = true;
Prince.registerPostLayoutFunc(handleMultipage(1));

const headerTempl = document.getElementById("mp-header-templ");
const footerTempl = document.getElementById("mp-footer-templ");

// handleMultipage detects page breaks and attaches the header and footer templates
// for each page with the subtotal amount
function handleMultipage(currPage) {
  return function () {
    const lines = document.querySelectorAll("tr[data-subtotal]");
    if (lines.length == 0) return;

    let subtotal = "";

    // iterate the lines, detect page breaks and handle them
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      const boxes = line.getPrinceBoxes();
      const page = boxes[0].pageNum

      subtotal = line.getAttribute("data-subtotal");

      if (page > currPage) {
        handlePageBreak(currPage, subtotal);
        return;
      }
    }

    const totals = document.querySelector(".tail");
    const boxes = totals.getPrinceBoxes();
    const tailPage = boxes[0].pageNum;

    // keep handling page breaks to the tail page
    if (currPage < tailPage) {
      handlePageBreak(currPage, subtotal);
      return;
    }
  }
}

function handlePageBreak(page, amount) {
  attach(headerTempl, page + 1, amount);
  attach(footerTempl, page, amount);

  // Since we have changed the DOM, we need to register a new function to handle the
  // next page breaks after Prince refreshes the layout with the changes.
  Prince.registerPostLayoutFunc(handleMultipage(page + 1));
}

function attach(template, page, amount) {
  var body = document.getElementsByTagName('body')[0];
  var el = template.cloneNode(true);
  el.id = "";
  el.setAttribute("style", "-prince-float-defer-page: " + (page - 1));
  el.querySelector(".amount").innerHTML = amount;
  el.style.display = "block";

  body.insertBefore(el, body.firstChild);
}
