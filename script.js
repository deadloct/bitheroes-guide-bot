async function getJSON() {
    const response = await fetch("commands.json");
    const categories = await response.json();
    return categories;
}

function fams(guide) {
    if (guide.fams && guide.fams.length) {
        return `<div><em>Fams:</em> ${guide.fams.join(", ")}</div>`;
    }

    return "";
}

function builds(guide) {
    if (guide.builds && guide.builds.length) {
        return `<div><em>Builds:</em> ${guide.builds.join(", ")}</div>`;
    }

    return "";
}

function linkText(link) {
    const MAX_LENGTH = 30;
    if (link.length < MAX_LENGTH) {
        return link
    }

    return `${link.substring(0, MAX_LENGTH)}...`;
}

function attachment(item) {
    switch (item.attachmenttype) {
        case "file":
            return `<li class="attachment-item"><a href="responses/${item.filename}" target="_BLANK">${item.filename} (${item.contenttype})</a></li>`;
        case "markdown":
            return `<li class="attachment-item"><a href="responses/${item.filename}" target="_BLANK">${item.filename} (markdown/text)</a></li>`;
        case "link":
            return `<li class="attachment-item"><a href="${item.link}" target="_BLANK">${linkText(item.link)}</a></li>`;
    }
}

function attachments(guide) {
    if (!guide.attachments) {
        return ""
    }

    return `<ul>${guide.attachments.map(attachment).join("")}</ul>`
}

function renderGuide(guide) {
    return `
        <li class="guide-item">
            <div class="guide-name">${guide.name}</div>
            ${fams(guide)}
            ${builds(guide)}
            ${attachments(guide)} 
        </li>
    `;
}

function categoryName(category) {
    return category.name.replace("guides-", "").replace("-", " ").trim();

}

function renderCategory(category) {
    const items = category.guides.map(renderGuide).join("");
    return `
        <h2>${categoryName(category)}</h2>
        <div class="category-description">${category.description}</div>
        <ul>${items}</ul>
    `;
}

function renderCategories(categories) {
    return categories.map(renderCategory).join("");
}

function renderResults(html) {
    document.getElementById("results").innerHTML = html;
}

async function run() {
    try {
        const categories = await getJSON();
        if (categories.length == 0) {
            throw new Error("no guide categories found")
        }

        renderResults(renderCategories(categories));
    } catch(err) {
        renderResults(err);
    }
}

document.addEventListener("DOMContentLoaded", e => {
    run();    
});
