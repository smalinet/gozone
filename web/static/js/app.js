// gozone - PowerDNS Admin Interface
console.log('gozone - PowerDNS Admin Interface');

(function() {
    var theme = localStorage.getItem('gozone-theme') || 'light';
    document.documentElement.setAttribute('data-theme', theme);

    var collapsed = localStorage.getItem('gozone-sidebar') === 'true';
    if (collapsed) {
        document.body.classList.add('sidebar-collapsed');
    }
})();

function toggleTheme() {
    var current = document.documentElement.getAttribute('data-theme');
    var next = current === 'dark' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', next);
    localStorage.setItem('gozone-theme', next);
}

function toggleSidebar() {
    document.body.classList.toggle('sidebar-collapsed');
    var collapsed = document.body.classList.contains('sidebar-collapsed');
    localStorage.setItem('gozone-sidebar', collapsed);
}
