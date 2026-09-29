// site.js — the feedback layer on top of htmx: toasts, a styled confirm
// dialog, flash messages that survive a redirect, and the editor's autosave
// state. Loaded with defer after htmx, so `htmx` is defined by now.
(() => {
	"use strict";

	/* ---------- toasts ---------- */

	const toasts = document.getElementById("toasts");

	function toast(message, kind = "info", ms = 3200) {
		if (!toasts || !message) return;
		const el = document.createElement("div");
		el.className = `toast toast--${kind}`;
		el.setAttribute("role", kind === "error" ? "alert" : "status");
		el.textContent = message;
		el.addEventListener("click", () => dismiss(el));
		toasts.append(el);
		// errors stay until clicked-ish: give them longer to be read
		setTimeout(() => dismiss(el), kind === "error" ? ms * 2 : ms);
	}

	function dismiss(el) {
		if (el.dataset.leaving) return;
		el.dataset.leaving = "true";
		el.addEventListener("animationend", () => el.remove(), { once: true });
		// reduced motion: animation is ~1ms, but guard against it never firing
		setTimeout(() => el.remove(), 400);
	}

	/* ---------- flash across redirects ---------- */

	// buttons with data-flash queue a message on success; it is shown on
	// whatever page the server redirects to.
	const FLASH_KEY = "flash";

	function readFlash() {
		try {
			const msg = sessionStorage.getItem(FLASH_KEY);
			if (msg) {
				sessionStorage.removeItem(FLASH_KEY);
				toast(msg, "success");
			}
		} catch {}
	}

	function writeFlash(msg) {
		try {
			sessionStorage.setItem(FLASH_KEY, msg);
		} catch {}
	}

	readFlash();

	/* ---------- confirm dialog ---------- */

	const dialog = document.getElementById("confirm");

	document.addEventListener("htmx:confirm", (evt) => {
		const question = evt.detail.question;
		if (!question || !dialog?.showModal) return; // no hx-confirm, or no <dialog>: htmx default
		evt.preventDefault();

		dialog.querySelector(".confirm__msg").textContent = question;
		const ok = dialog.querySelector('[value="ok"]');
		const danger = evt.detail.elt.classList.contains("admin-action--danger");
		ok.classList.toggle("admin-action--danger-solid", danger);
		ok.textContent = evt.detail.elt.textContent.trim() || "confirm";

		dialog.returnValue = "";
		dialog.showModal();
		ok.focus();
		dialog.addEventListener(
			"close",
			() => {
				if (dialog.returnValue === "ok") evt.detail.issueRequest(true);
			},
			{ once: true },
		);
	});

	// click on the backdrop closes it
	dialog?.addEventListener("click", (e) => {
		if (e.target === dialog) dialog.close("cancel");
	});

	/* ---------- request errors ---------- */

	function errorText(xhr) {
		const body = (xhr?.responseText || "").trim();
		if (body && body.length < 200) return body;
		return xhr?.status ? `request failed (${xhr.status})` : "request failed";
	}

	document.addEventListener("htmx:responseError", (evt) => {
		toast(errorText(evt.detail.xhr), "error");
	});

	document.addEventListener("htmx:sendError", () => {
		toast("can't reach the server — check your connection", "error");
	});

	document.addEventListener("htmx:afterRequest", (evt) => {
		const msg = evt.detail.elt?.dataset?.flash;
		if (msg && evt.detail.successful) writeFlash(msg);
	});

	/* ---------- editor autosave ---------- */

	const saveState = document.getElementById("save-state");
	if (!saveState) return; // not the editor

	// field -> "dirty" | "saving" | "error"; absent means saved
	const pending = new Map();
	const saved = new Map();
	const inflight = new Map(); // field -> value being sent
	const fieldEl = (name) => document.querySelector(`[data-field="${name}"]`);
	const statusEl = (name) => document.querySelector(`[data-status-for="${name}"]`);

	for (const el of document.querySelectorAll("[data-field]")) {
		saved.set(el.dataset.field, el.value);
	}

	function setField(name, state, text) {
		if (state === "saved") pending.delete(name);
		else pending.set(name, state);

		const s = statusEl(name);
		if (s) {
			s.dataset.state = state;
			s.textContent = text ?? { dirty: "unsaved", saving: "saving…", saved: "saved ✓", error: "not saved" }[state];
			// retrigger the pop animation
			s.classList.remove("is-fresh");
			void s.offsetWidth;
			s.classList.add("is-fresh");
		}
		renderSaveState();
	}

	function renderSaveState() {
		const states = [...pending.values()];
		let state = "idle";
		let text = "all changes saved";
		if (states.includes("error")) {
			state = "error";
			text = "some changes not saved";
		} else if (states.includes("saving")) {
			state = "saving";
			text = "saving…";
		} else if (states.includes("dirty")) {
			state = "dirty";
			text = "unsaved changes";
		}
		if (saveState.dataset.state === state) return;
		saveState.dataset.state = state;
		saveState.textContent = text;
	}

	const fieldOf = (elt) => elt?.closest?.("[data-field]")?.dataset.field;

	document.addEventListener("input", (e) => {
		const name = fieldOf(e.target);
		if (!name) return;
		const el = fieldEl(name);
		if (el.value === saved.get(name)) {
			if (pending.get(name) === "dirty") setField(name, "saved", "");
		} else if (pending.get(name) !== "dirty") {
			setField(name, "dirty");
		}
	});

	document.addEventListener("htmx:beforeRequest", (evt) => {
		const name = fieldOf(evt.detail.elt);
		if (!name) return;
		inflight.set(name, fieldEl(name).value);
		setField(name, "saving");
	});

	document.addEventListener("htmx:afterRequest", (evt) => {
		const name = fieldOf(evt.detail.elt);
		if (!name) return;
		const xhr = evt.detail.xhr;
		if (evt.detail.successful) {
			saved.set(name, inflight.get(name) ?? fieldEl(name).value);
			// typed more while this request was in flight: still dirty
			if (fieldEl(name).value !== saved.get(name)) setField(name, "dirty");
			else setField(name, "saved");
		} else if (xhr && xhr.status === 0 && !evt.detail.failed) {
			// aborted by hx-sync replace; the newer request reports for it
		} else {
			setField(name, "error", errorText(xhr));
		}
	});

	// the slug comes back out-of-band after a title save: that value is the saved one
	document.addEventListener("htmx:oobAfterSwap", (evt) => {
		const name = fieldOf(evt.detail.target || evt.target);
		if (!name) return;
		const el = fieldEl(name);
		if (saved.get(name) === el.value) return;
		saved.set(name, el.value);
		setField(name, "saved", "updated from title");
		el.classList.remove("is-updated");
		void el.offsetWidth;
		el.classList.add("is-updated");
	});

	// ctrl/cmd+s: flush everything that hasn't saved yet
	document.addEventListener("keydown", (e) => {
		if (!(e.ctrlKey || e.metaKey) || e.key.toLowerCase() !== "s") return;
		e.preventDefault();
		const names = [...pending].filter(([, s]) => s !== "saving").map(([n]) => n);
		if (names.length === 0) {
			toast("nothing to save", "info", 1600);
			return;
		}
		for (const n of names) htmx.trigger(fieldEl(n), "save-now");
	});

	window.addEventListener("beforeunload", (e) => {
		if (pending.size > 0) e.preventDefault();
	});

	// description character counter
	for (const el of document.querySelectorAll("[data-counter]")) {
		const out = document.getElementById(el.dataset.counter);
		const max = Number(el.getAttribute("maxlength")) || Infinity;
		const render = () => {
			out.textContent = el.value.length;
			out.parentElement.classList.toggle("is-near-limit", el.value.length > max * 0.9);
		};
		el.addEventListener("input", render);
		render();
	}
})();
