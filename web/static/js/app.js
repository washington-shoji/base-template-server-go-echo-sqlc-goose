(function () {
  function toggleTheme() {
    var root = document.documentElement;
    var next = root.classList.contains("dark") ? "light" : "dark";
    root.classList.toggle("dark", next === "dark");
    try {
      localStorage.setItem("theme", next);
    } catch (e) {}
  }

  document.addEventListener("click", function (ev) {
    var t = ev.target;
    if (!(t instanceof Element)) return;

    if (t.closest("[data-theme-toggle]")) {
      ev.preventDefault();
      toggleTheme();
      return;
    }

    var openBtn = t.closest("[data-dialog-open]");
    if (openBtn) {
      var id = openBtn.getAttribute("data-dialog-open");
      var dlg = id && document.getElementById(id);
      if (dlg && typeof dlg.showModal === "function") {
        var url = openBtn.getAttribute("data-delete-url");
        if (url) {
          dlg.setAttribute("data-confirm-action", url);
        }
        dlg.showModal();
      }
      return;
    }

    var closeBtn = t.closest("[data-dialog-close]");
    if (closeBtn) {
      var closeId = closeBtn.getAttribute("data-dialog-close");
      var closeDlg = closeId && document.getElementById(closeId);
      if (closeDlg && typeof closeDlg.close === "function") {
        closeDlg.close();
      }
      return;
    }

    var confirmBtn = t.closest("[data-dialog-confirm]");
    if (confirmBtn) {
      var confirmId = confirmBtn.getAttribute("data-dialog-confirm");
      var confirmDlg = confirmId && document.getElementById(confirmId);
      var action = confirmDlg && confirmDlg.getAttribute("data-confirm-action");
      var target = (confirmDlg && confirmDlg.getAttribute("data-confirm-target")) || "#todo-list";
      var csrf = (confirmDlg && confirmDlg.getAttribute("data-csrf")) || "";
      if (confirmDlg && typeof confirmDlg.close === "function") {
        confirmDlg.close();
      }
      if (action && window.htmx) {
        window.htmx.ajax("DELETE", action, {
          target: target,
          swap: "innerHTML",
          headers: { "X-CSRF-Token": csrf },
        });
      }
    }
  });

  document.querySelectorAll("[data-flash]").forEach(function (el) {
    setTimeout(function () {
      el.style.transition = "opacity 200ms";
      el.style.opacity = "0";
      setTimeout(function () {
        el.remove();
      }, 220);
    }, 5000);
  });
})();
