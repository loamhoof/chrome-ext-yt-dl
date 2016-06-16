;(() => {
    const TARGETS = [
    ]

    chrome.commands.onCommand.addListener((command) => {
        chrome.tabs.query({url: 'https://www.youtube.com/*'}, (tabs) => {
            for (let tab of tabs) {
                for (let target of TARGETS) {
                    let xhr = new XMLHttpRequest();
                    xhr.open('GET', `${target}/${encodeURIComponent(tab.url)}`);
                    xhr.send();
                }
                chrome.extension.getBackgroundPage().console.log(tab.url);
            }
        })
    });
})();
