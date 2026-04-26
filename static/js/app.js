function copyDraft() {
    const draftText = document.querySelector(".whitespace-pre-wrap");
    if (draftText) {
        navigator.clipboard.writeText(draftText.innerText).then(() => {
            alert("Draft copied to clipboard!");
        });
    }
}
