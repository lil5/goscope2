window.dayjs.extend(window.dayjs_plugin_calendar);
window.dayjs.extend(window.dayjs_plugin_updateLocale);
window.dayjs.extend(window.dayjs_plugin_relativeTime, {
  thresholds: [
    { l: "s", r: 1 },
    { l: "ss", r: 59, d: "second" },
    { l: "m", r: 1 },
    { l: "mm", r: 59, d: "minute" },
    { l: "h", r: 1 },
    { l: "hh", r: 23, d: "hour" },
    { l: "d", r: 1 },
    { l: "dd", r: 29, d: "day" },
    { l: "M", r: 1 },
    { l: "MM", r: 11, d: "month" },
    { l: "y", r: 1 },
    { l: "yy", d: "year" },
  ],
});
window.dayjs.updateLocale("en", {
  relativeTime: {
    future: "in %s",
    past: "%s",
    s: "a second",
    ss: "%d seconds",
    m: "a minute",
    mm: "%d minutes",
    h: "an hour",
    hh: "%d hours",
    d: "a day",
    dd: "%d days",
    M: "a month",
    MM: "%d months",
    y: "a year",
    yy: "%d years",
  },
});

async function fetchLogs(t, page) {
  const res = await fetch(`./api?type=${t}&page=${page}`);
  return await res.json().catch((err) => {
    console.error(err);
    return [];
  });
}

document.addEventListener("alpine:init", () => {
  window.Alpine.data("body", () => ({
    logs: [],
    logsFiltered: [],
    apiPage: 1,
    page: 20,
    pageEnd: false,
    filter: {
      type: "http",
      status: ["1", "2", "3", "4", "5"],
      message: "",
    },
    selectedLog: null,
    darkMode: null,
    async init() {
      if (
        window.matchMedia &&
        window.matchMedia("(prefers-color-scheme: dark)").matches
      ) {
        this.darkMode = true;
      }
      this.logs = await fetchLogs(this.filter.type, 1);
      await this.setList();
    },
    async selectType(id) {
      this.filter = {
        type: id,
        status: ["1", "2", "3", "4", "5"],
        message: "",
      };
      this.selectedLog = null;
      this.apiPage = 1;
      this.page = 20;
      this.pageEnd = false;
      this.logs = await fetchLogs(this.filter.type, this.apiPage);
      await this.setList();
      document.getElementById("my-drawer").checked = true;
    },
    async selectFilter() {
      this.selectedLog = null;
      this.page = 20;
      this.pageEnd = false;
      await this.setList();
    },
    selectLog(item) {
      this.selectedLog = item;
      document.getElementById("my-drawer").checked = false;
    },
    async setList() {
      if (!this.logs.length) {
        this.logsFiltered = [];
        this.pageEnd = true;
        return;
      }
      const { type, status, message } = this.filter;
      console.log(type, status, message);
      let i = 0,
        arr = [];
      while (!this.pageEnd && arr.length < this.page) {
        let item = this.logs[i];
        if (!item) throw "item is undefined on index: " + i;
        if (item.type == type) {
          let ok = true;
          if (
            type == "http" &&
            status.length &&
            item.status &&
            !status.includes((item.status + "0").charAt(0))
          ) {
            ok = false;
          }
          if (message && !item.message.includes(message)) {
            ok = false;
          }
          if (ok) arr.push(item);
        }

        i++;
        if (i >= this.logs.length) {
          let isEnd = true;
          try {
            const res = await fetchLogs(type, ++this.apiPage);
            if (Array.isArray(res) && res.length) {
              console.log("page", this.apiPage);
              this.logs.push(res);
              isEnd = false;
            }
          } catch (err) {
            console.error(err);
          }
          this.pageEnd = isEnd;
        }
      }
      this.logsFiltered = arr;
    },
    next() {
      this.page += 10;
      this.setList();
    },
    handleScroll(e) {
      if (!this.pageEnd && this.isAtBottom(e.target)) {
        this.$dispatch("scrolled-to-bottom");
      }
    },
    // utils
    day(timestamp) {
      const d = window.dayjs(timestamp);
      const now = window.dayjs();

      if (now.diff(d, "day") < 1) {
        return d.from(now, true);
      }

      return d.calendar(null, {
        nextWeek: "YYYY-MM-DD",
        nextDay: "YYYY-MM-DD",
        sameDay: "HH:mm ss",
        lastDay: "ddd HH:mm",
        lastWeek: "ddd HH:mm",
        sameElse: "YYYY-MM-DD",
      });
    },
    isAtBottom(el) {
      let sh = el.scrollHeight,
        st = el.scrollTop,
        ht = el.offsetHeight;
      if (ht == 0) return true;
      return st == sh - ht;
    },
    async countOccurrence(hash) {
      if (!hash) return "1";
      let i = 0,
        count = 0;
      const dataLogsLen = this.logs.length;
      while (i < dataLogsLen && count < 100) {
        let item = this.logs[i];
        if (item.hash == hash) count++;
        i++;
      }
      if (count === 100) return "99+";
      return "" + count;
    },
  }));
});
