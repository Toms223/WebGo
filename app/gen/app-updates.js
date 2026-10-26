(() => {
  if (!("serviceWorker" in navigator)) {
    return;
  }

  const autoReload = {{.AutoUpdate}};
  const workers = navigator.serviceWorker;
  let controller = workers.controller;
  let controllerReplaced = false;
  let reloading = false;
  let checking = false;

  window.goappReloadForUpdate = () => {
    if (autoReload && controllerReplaced && !reloading) {
      reloading = true;
      window.location.reload();
    }
  };

  workers.addEventListener("controllerchange", () => {
    const previous = controller;
    controller = workers.controller;
    if (previous && controller && previous !== controller) {
      controllerReplaced = true;
      window.goappReloadForUpdate();
    }
  });

  async function checkForUpdate() {
    if (checking || navigator.onLine === false ||
        document.visibilityState === "hidden") {
      return;
    }
    checking = true;
    try {
      const registration = await workers.getRegistration();
      if (registration) {
        await registration.update();
      }
    } catch (error) {
      console.debug("WebGo update check deferred:", error);
    } finally {
      checking = false;
    }
  }

  window.addEventListener("load", checkForUpdate);
  window.addEventListener("online", checkForUpdate);
  document.addEventListener("visibilitychange", checkForUpdate);
  window.setInterval(checkForUpdate, 60_000);
})();
