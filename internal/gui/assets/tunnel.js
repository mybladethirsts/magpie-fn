// The Cloud tunnel block on Settings → Network and sharing (magpie-fn).
// A quick tunnel needs nothing but the Start button and gives a temporary
// trycloudflare.com URL; a named tunnel runs one created in the Cloudflare
// dashboard (Zero Trust → Networks → Tunnels → Create) with its token, and
// keeps the public hostname you chose there across restarts.
(() => {
	const box = $("#tunnelList");
	if (!box) return;

	const api = async (path, body) => {
		const r = await fetch(path, body ? {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(body),
		} : undefined);
		const j = await r.json().catch(() => ({}));
		if (!r.ok) throw new Error(j.error || ("HTTP " + r.status));
		return j;
	};

	let data = null; // {state, config}

	const noteFor = (c, s) => {
		if (c.mode === "named") return s.running
			? "Named tunnel is up. Your public hostname is the one you set in the Cloudflare dashboard."
			: "Paste the tunnel token from Cloudflare (Zero Trust → Networks → Tunnels), then Start. Create the tunnel with its service set to http://localhost:" + (c.target === "web" ? "3430" : "3425") + ".";
		return s.running
			? "The quick URL above is alive while this tunnel runs and changes on every restart."
			: "Start gives a temporary https://….trycloudflare.com URL — zero configuration.";
	};

	const render = () => {
		const s = data?.state || { running: false, mode: "quick", target: "gateway" };
		const c = data?.config || { mode: "quick", target: "gateway", tokenSet: false };
		box.querySelectorAll("[data-mode]").forEach(b => b.classList.toggle("on", b.dataset.mode === c.mode));
		box.querySelectorAll("[data-target]").forEach(b => b.classList.toggle("on", b.dataset.target === c.target));
		box.querySelector("#tunnelTokenRow").hidden = c.mode !== "named";
		const run = box.querySelector("#tunnelRun");
		run.disabled = s.running;
		run.textContent = s.running ? "Restart with new choices" : "Start tunnel";
		box.querySelector("#tunnelStop").disabled = !s.running;
		const status = box.querySelector("#tunnelStatus");
		status.classList.toggle("on", s.running);
		status.textContent = s.running ? "Running" : "Stopped";
		const url = box.querySelector("#tunnelURL");
		const copy = box.querySelector("#tunnelURLCopy");
		if (s.url) {
			url.href = s.url;
			url.textContent = s.url;
			url.hidden = copy.hidden = false;
		} else {
			url.hidden = copy.hidden = true;
		}
		box.querySelector("#tunnelNote").textContent = s.error || noteFor(c, s);
	};

	const refresh = async () => {
		try {
			data = await api("/api/tunnel");
			render();
		} catch (e) {
			box.querySelector("#tunnelStatus").textContent = "Unavailable";
			box.querySelector("#tunnelNote").textContent = e.message;
		}
	};

	box.querySelectorAll("[data-mode]").forEach(b => b.addEventListener("click", () => {
		if (!data) return;
		data.config.mode = b.dataset.mode;
		render();
	}));
	box.querySelectorAll("[data-target]").forEach(b => b.addEventListener("click", () => {
		if (!data) return;
		data.config.target = b.dataset.target;
		render();
	}));

	box.querySelector("#tunnelRun").addEventListener("click", async () => {
		const run = box.querySelector("#tunnelRun");
		run.disabled = true;
		try {
			data = await api("/api/tunnel/start", {
				mode: data.config.mode,
				target: data.config.target,
				token: box.querySelector("#tunnelToken").value.trim(),
			});
			render();
		} catch (e) {
			box.querySelector("#tunnelNote").textContent = e.message;
			run.disabled = false;
		}
	});
	box.querySelector("#tunnelStop").addEventListener("click", async () => {
		try {
			data = await api("/api/tunnel/stop");
			render();
		} catch (e) {
			box.querySelector("#tunnelNote").textContent = e.message;
		}
	});
	box.querySelector("#tunnelURLCopy").addEventListener("click", async () => {
		const u = box.querySelector("#tunnelURL").textContent;
		try { await navigator.clipboard.writeText(u); } catch { /* desktop: nothing more to do */ }
		box.querySelector("#tunnelNote").textContent = "Copied";
	});

	// the quick URL appears a moment after Start; keep the card current
	setInterval(refresh, 5000);
	refresh();
})();
