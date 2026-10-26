const cacheName = "app-" + "{{.Version}}";
const resourcesToCache = {{.ResourcesToCache}};
const appResources = new Set(resourcesToCache
  .map((resource) => new URL(resource, self.location.origin))
  .filter((url) => url.origin === self.location.origin)
  .map((url) => url.href));

self.addEventListener("install", (event) => {
  event.waitUntil(installWorker());
});

async function installWorker() {
  const cache = await caches.open(cacheName);
  try {
    await cache.addAll([...appResources].map((url) => new Request(url, {
      cache: "reload",
      credentials: "same-origin",
      redirect: "error",
    })));
  } catch (error) {
    await caches.delete(cacheName);
    throw error;
  }
  await self.skipWaiting();
}

self.addEventListener("activate", (event) => {
  event.waitUntil(activateWorker());
});

async function activateWorker() {
  const keys = await caches.keys();
  await Promise.all(keys
    .filter((key) => key.startsWith("app-") && key !== cacheName)
    .map((key) => caches.delete(key)));
  await self.clients.claim();
}

self.addEventListener("fetch", (event) => {
  const request = event.request;
  if (request.method !== "GET" || !appResources.has(request.url)) {
    return;
  }
  event.respondWith(fetchAppResource(request));
});

async function fetchAppResource(request) {
  const cache = await caches.open(cacheName);
  const cached = await cache.match(request);
  return cached || fetch(request);
}

self.addEventListener("push", (event) => {
  event.waitUntil((async () => {
    let notification;

    try {
      notification = event.data ? event.data.json() : null;
    } catch {
      notification = null;
    }

    if (!notification) {
      return;
    }

    await showNotification(self.registration, notification);
  })());
});

self.addEventListener("message", (event) => {
  const msg = event.data;
  if (!msg || msg.type !== "goapp:notify") {
    return;
  }

  event.waitUntil(
    showNotification(self.registration, msg.options)
  );
});

async function showNotification(registration, notification) {
  const title = notification.title || "Notification";

  let actions = [];
  for (let i in notification.actions) {
    const action = notification.actions[i];
    actions.push({
      action: action.action,
      path: action.path,
    });
    delete action.path;
  }

  await registration.showNotification(title, {
    body: notification.body,
    icon: notification.icon,
    badge: notification.badge,
    actions: notification.actions,
    data: {
      goapp: {
        path: notification.path,
        actions: actions
      }
    }
  });
}

self.addEventListener("notificationclick", (event) => {
  event.notification.close();

  const notification = event.notification;
  let path = notification.data.goapp.path;

  for (let i in notification.data.goapp.actions) {
    const action = notification.data.goapp.actions[i];
    if (action.action === event.action) {
      path = action.path;
      break;
    }
  }

  event.waitUntil(
    clients
      .matchAll({
        type: "window",
      })
      .then((clientList) => {
        for (var i = 0; i < clientList.length; i++) {
          let client = clientList[i];
          if ("focus" in client) {
            client.focus();
            client.postMessage({
              goapp: {
                type: "notification",
                path: path,
              },
            });
            return;
          }
        }

        if (clients.openWindow) {
          return clients.openWindow(path);
        }
      })
  );
});