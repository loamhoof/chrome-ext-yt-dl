;(() => {
    chrome.commands.onCommand.addListener((command) => {
        chrome.tabs.query({url: 'https://www.youtube.com/*'}, (tabs) => {
            for (let tab of tabs) {
                let xhr = new XMLHttpRequest();
                xhr.open('GET', `${chrome.runtime.getManifest().settings.target}/${encodeURIComponent(tab.url)}`);
                xhr.send();
            }
        })
    });
})();
