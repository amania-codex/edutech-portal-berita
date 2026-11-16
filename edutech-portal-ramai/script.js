const $ = (selector, scope = document) => scope.querySelector(selector);
const $$ = (selector, scope = document) => [...scope.querySelectorAll(selector)];

const dateNode = $("#todayDate");
if (dateNode) {
  const now = new Date();
  dateNode.textContent = new Intl.DateTimeFormat("id-ID", {
    weekday: "long", day: "numeric", month: "long", year: "numeric"
  }).format(now);
}

const searchPanel = $("#searchPanel");
const searchInput = $("#searchInput");
const noResults = $("#noResults");
const contentSections = $$(".news-section, .hero-grid, .topic-strip, .event-section, .callpapers");
$("#searchToggle")?.addEventListener("click", () => {
  searchPanel.classList.toggle("open");
  if (searchPanel.classList.contains("open")) searchInput.focus();
});
$("#searchClose")?.addEventListener("click", () => {
  searchPanel.classList.remove("open");
  searchInput.value = "";
  filterStories("");
});
searchInput?.addEventListener("input", event => filterStories(event.target.value));

function filterStories(query) {
  const q = query.trim().toLocaleLowerCase("id");
  let visible = 0;
  $$(".story.searchable").forEach(story => {
    const matches = !q || story.textContent.toLocaleLowerCase("id").includes(q);
    story.hidden = !matches;
    if (matches) visible++;
  });
  contentSections.forEach(section => {
    const visibleStories = $$(".story.searchable:not([hidden])", section);
    if (section.querySelector(".story.searchable")) {
      section.hidden = visibleStories.length === 0;
    }
  });
  if (noResults) noResults.hidden = !q || visible > 0;
}

$("#themeToggle")?.addEventListener("click", () => {
  document.body.classList.toggle("dark");
  const isDark = document.body.classList.contains("dark");
  $("#themeToggle").textContent = isDark ? "☀" : "☾";
  showToast(isDark ? "Mode gelap diaktifkan" : "Mode terang diaktifkan");
});

$("#mobileMenu")?.addEventListener("click", () => {
  const nav = $("#mainNav");
  nav.classList.toggle("mobile-open");
  $("#mobileMenu").textContent = nav.classList.contains("mobile-open") ? "✕" : "☰";
});
$$(".nav-inner a").forEach(link => link.addEventListener("click", () => {
  $("#mainNav").classList.remove("mobile-open");
  $("#mobileMenu").textContent = "☰";
}));

const popularTrack = $("#popularTrack");
$("#popularPrev")?.addEventListener("click", () => popularTrack.scrollBy({left: -popularTrack.clientWidth * .75, behavior: "smooth"}));
$("#popularNext")?.addEventListener("click", () => popularTrack.scrollBy({left: popularTrack.clientWidth * .75, behavior: "smooth"}));

const toast = $("#toast");
let toastTimer;
function showToast(message) {
  toast.textContent = message;
  toast.classList.add("show");
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => toast.classList.remove("show"), 2600);
}

$("#newsletterForm")?.addEventListener("submit", event => {
  event.preventDefault();
  const email = $("#emailInput").value.trim();
  if (!email) return;
  $("#newsletterMessage").textContent = "Terima kasih! Simulasi berlangganan berhasil.";
  $("#emailInput").value = "";
  showToast("Terima kasih sudah berlangganan!");
});

const backTop = $("#backTop");
window.addEventListener("scroll", () => {
  backTop.classList.toggle("visible", window.scrollY > 500);
}, {passive: true});
backTop?.addEventListener("click", () => window.scrollTo({top: 0, behavior: "smooth"}));

$$('a[href="#detail"], a[href="#daftar"], a[href="#submit"], a[href="#kalender"]').forEach(link => {
  link.addEventListener("click", event => {
    event.preventDefault();
    showToast("Ini masih demo. Hubungkan ke halaman detail atau formulir backend.");
  });
});
