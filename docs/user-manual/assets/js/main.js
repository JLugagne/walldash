(function () {
  'use strict';

  var root = document.documentElement;

  function effectiveTheme() {
    var stored = localStorage.getItem('walldash-theme');
    if (stored === 'light' || stored === 'dark') return stored;
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }

  var themeToggle = document.querySelector('[data-theme-toggle]');
  if (themeToggle) {
    themeToggle.addEventListener('click', function () {
      var next = effectiveTheme() === 'dark' ? 'light' : 'dark';
      root.setAttribute('data-theme', next);
      try {
        localStorage.setItem('walldash-theme', next);
      } catch (e) {}
    });
  }

  var menuToggle = document.querySelector('[data-menu-toggle]');
  var scrim = document.querySelector('.scrim');
  if (menuToggle) {
    var setOpen = function (open) {
      document.body.classList.toggle('nav-open', open);
      menuToggle.setAttribute('aria-expanded', String(open));
    };
    menuToggle.addEventListener('click', function () {
      setOpen(!document.body.classList.contains('nav-open'));
    });
    if (scrim) scrim.addEventListener('click', function () { setOpen(false); });
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape') setOpen(false);
    });
  }
})();
