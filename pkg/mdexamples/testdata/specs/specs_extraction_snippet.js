function getSection(el) {
  let headers = new Array(7)
  while (el){
    if (el.nodeName.length === 2 && el.nodeName.startsWith("H")){
      let h;
      try {
        h = parseInt(el.nodeName[1], 10);
        if (h<1 || h>6){
          throw new Error("incorrect header");
        }
      } catch (error) {
        continue
      }
      if (!headers[h]){
        headers[h] = el.textContent
        if (h==1) {
          break
        }
      }
    }
    if (el.previousElementSibling) {
      el = el.previousElementSibling
    }else{
      el = el.parentElement
    }
  }
  return headers.filter((v)=>v).map((v)=>`[${v.replace(/(^\s*[0-9.]*\s*)|(\s*$)/g, "")}]`).join("");
}

// Extract examples from markdown spec html page
// GFM actually doesn't have an up-to-date non-html version of the spec
// so we extract the examples from a web page
const data = [...document.querySelectorAll("div.example")].map((ex, idx) => {
  const txt = [...ex.querySelectorAll("div.column>pre")].map((item) => item.textContent.replaceAll("→", "\t"));
  return {
    markdown: txt[0],
    html: txt[1],
    id: idx+1,
    section: getSection(ex),
  };
});

const url = new URL(window.location)
url.hash = "example"
copy(JSON.stringify({
  exampleLinkFormat: url.toString()+"-%d",
  examples: data
}))


