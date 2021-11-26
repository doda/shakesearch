const Controller = {
  search: (ev) => {
    ev.preventDefault();
    const form = document.getElementById("form");
    const data = Object.fromEntries(new FormData(form));
    const response = fetch(`/search?q=${data.query}`).then((response) => {
      response.json().then((results) => {
        Controller.updateTable(results);
      });
    });
  },

  updateTable: (results) => {
    const table = document.getElementById("table-body");
    const rows = [];
    rows.push(`<tr>${results.length} results</tr>`);
    for (let result of results) {
      rows.push(`<tr><h4>${result.title}</h4></tr>`);
      for (let line of result.lines) {
        rows.push(`<tr>${line}</tr>`);
      } 
    }
    table.innerHTML = rows.join("");
  },
};

const form = document.getElementById("form");
form.addEventListener("submit", Controller.search);
