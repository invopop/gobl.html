Prince.trackBoxes = true;
Prince.registerPostLayoutFunc(handleMultipage(1));

const headerTemplate = document.getElementById("mp-header-templ");
const footerTemplate = document.getElementById("mp-footer-templ");

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

      if (page > currPage) {
        handlePageBreak(currPage, subtotal);
        return;
      }

      subtotal = line.getAttribute("data-subtotal");
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
  const headerInserted = upsertTemplate(headerTemplate, page + 1, amount);
  const footerInserted = upsertTemplate(footerTemplate, page, amount);

  if (headerInserted || footerInserted) {
    // the inserted elements may have pushed lines to the next page, we need to
    // register a new function to handle the same page again after Prince
    // refreshes the layout with the changes
    Prince.registerPostLayoutFunc(handleMultipage(page));
  } else {
    // no lines pushed, so we process the next page
    Prince.registerPostLayoutFunc(handleMultipage(page + 1));
  }
}

function upsertTemplate(template, page, amount) {
  let inserted = false;
  const id = template.id.replace("-templ", "-" + page);
  let el = document.getElementById(id);

  if (!el) {
    // the element doesn't exist, insert it from the template
    el = template.cloneNode(true);
    el.id = id;
    el.setAttribute("style", "-prince-float-defer-page: " + (page - 1));
    el.style.display = "block";

    const body = document.getElementsByTagName('body')[0];
    body.insertBefore(el, body.firstChild);
    inserted = true;
  }

  // update the element based on the parameters; even if the element already
  // existed the parameters may have changed after a DOM change
  el.querySelector(".amount").innerHTML = amount;

  return inserted;
}
