export default {
  defaultBrowser: "Google Chrome",
  options: {
    logRequests: true,
  },
  handlers: [
    {
      match: "*newwindow*",
      browser: {
        name: "Google Chrome",
        newWindow: true
      }
    },
    {
      match: "*incognito*",
      browser: {
        name: "Google Chrome",
        incognito: true
      }
    },
    {
      match: "*profile*",
      browser: {
        name: "Google Chrome",
        profile: "Default",
        newWindow: true
      }
    },
    {
      match: "*combined*",
      browser: {
        name: "Google Chrome",
        newWindow: true,
        incognito: true,
        profile: "Default"
      }
    },
  ]
};
