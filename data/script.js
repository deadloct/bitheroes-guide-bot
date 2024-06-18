const MIN_TOKEN_LENGTH = 1;

class Search {
    constructor(categories) {
        this.index = {};
        this.guides = [];

        this.buildIndex(categories);
    }

    buildIndex(categories) {
        for (const cat of categories) {
            for (const guide of cat.guides) {
                let guideIndex = this.guides.length;
                guide.categoryName = cat.webname;
                this.guides.push(guide);

                const getValues = (object, parents = []) => Object.assign({}, ...Object
                    .entries(object)
                    .map(([k, v]) => v && typeof v === 'object'
                        ? getValues(v, [...parents, k])
                        : { [[...parents, k].join('.')]: v }
                    )
                );

                // From https://stackoverflow.com/a/34515563
                const searchable = Object.values(getValues(guide));
                const tokens = searchable
                    .join(" ")
                    .concat(" ", cat.webname, " ", cat.description)
                    .toLowerCase()
                    .replace(/[^a-zA-Z0-9]/g, " ")
                    .replace(/\s+/g, " ")
                    .trim()
                    .split(" ");

                for (const token of tokens) {
                    if (token.length < MIN_TOKEN_LENGTH) {
                        continue;
                    }

                    for (let i = MIN_TOKEN_LENGTH; i <= token.length; i++) {
                        const tokenVariant = token.substring(0, i);
                        if (tokenVariant in this.index) {
                            if (!this.index[tokenVariant].includes(guideIndex)) {
                                this.index[tokenVariant].push(guideIndex);
                            }
                        } else {
                            this.index[tokenVariant] = [guideIndex];
                        }
                    }
                }
            }
        }
    }

    Log() {
        console.log("Search index:", this.index);
        console.log("All possible search results:", this.guides);
    }

    Find(query) {
        query = query.replace(/[^a-zA-Z0-9\s]/g, " ").toLowerCase();
        const queries = query.trim().split(/\s+/);
        const results = [];
        let indices = new Set();

        for (let i = 0; i < queries.length; i++) {
            if (queries[i].length < MIN_TOKEN_LENGTH) {
                continue;
            }

            const matches = queries[i] in this.index ? this.index[queries[i]] : [];
            if (matches.length == 0) {
                indices = new Set();
                break;
            }

            if (indices.size == 0) {
                indices = new Set(matches);
                continue;
            }

            // Intersection of new and old
            indices = new Set(matches.filter(v => indices.has(v)));
        }
        
        for (const idx of indices) {
            results.push(this.guides[idx]);
        }

        console.log(`results for ${query}:`, indices, results);
        return results;
    }
}

const BuildUI = (() => {
    async function getJSON() {
        const response = await fetch("commands.json?202406190914");
        const categories = await response.json();
        return categories;
    }

    function obsolete(guide) {
        if (guide.obsolete && guide.obsolete.length) {
            return `
                <div class="obsolete">
                    <div class="obsolete-left"><i class="bi bi-x-circle-fill"></i></div>
                    <div class="obsolete-center"><strong>Obsolete</strong><br />${guide.obsolete}</div>
                    <div class="obsolete-right"><i class="bi bi-x-circle-fill"></i></div>
                </div>`;
        }

        return "";
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

    function attachment(item) {
        switch (item.attachmenttype) {
            case "file":
                return `<li class="attachment-item"><i class="bi bi-card-image"></i> <a href="responses/${item.filename}" target="_BLANK">${item.filename}</a> <span class="att-type">(${item.contenttype})</span></li>`;
            case "markdown":
                return `<li class="attachment-item"><i class="bi bi-file-earmark-text-fill"></i> <a href="responses/${item.filename}" target="_BLANK">${item.filename}</a> <span class="att-type">(markdown/text)</span></li>`;
            case "link":
                return `<li class="attachment-item"><i class="bi bi-box-arrow-up-right"></i> <a href="${item.link}" target="_BLANK">${item.link}</a> <span class="att-type">(external link)</span></li>`;
        }
    }

    function attachments(guide) {
        if (!guide.attachments) {
            return ""
        }

        return `<ul>${guide.attachments.map(attachment).join("")}</ul>`
    }

    function renderGuide(guide, isSearch) {
        let cat = "";
        if (isSearch && guide.categoryName) {
            cat = `<div>Category: ${guide.categoryName}</div>`
        }

        return `
            <li class="guide-item">
                <div class="guide-name">${guide.name}</div>
                ${obsolete(guide)}
                ${fams(guide)}
                ${builds(guide)}
                ${cat}
                ${attachments(guide)} 
            </li>
        `;
    }

    function categoryName(category) {
        return category.webname || category.name.replace("guides-", "").replace("-", " ").trim();
    }

    function categoryAnchor(category) {
        return category.isSearch ? 'search-results' : category.name;
    }

    function renderCategory(category) {
        const items = category.guides.map(g => renderGuide(g, category.isSearch)).join("");
        return `
            <h2 id="${categoryAnchor(category)}">${categoryName(category)}</h2>
            <div class="category-description">${category.description}</div>
            <ul>${items}</ul>
        `;
    }

    function renderTableOfContents(categories) {
        const items = categories
            .map(category => {
                const name = categoryName(category);
                const link = `#${categoryAnchor(category)}`
                return `<li><a href="${link}">${name}</a>`;
            })
            .join("");

        return `
            <div class="table-of-contents">
                <h2>Table of Contents</h2>
                <ol>${items}</ol>
            </div>
        `;
    }

    function renderCategories(categories) {
        categories.sort((a, b) => a.webname.localeCompare(b.webname));
        return [
            renderTableOfContents(categories),
            ...categories.map(renderCategory)
        ].join("");
    }

    function Render(target, html) {
        target.innerHTML = html;
    }

    let cachedFull;
    async function Full(target) {
        if (cachedFull && cachedFull.length > 0) {
            Render(target, cachedFull);
            return;
        }

        try {
            const categories = await getJSON();
            if (categories.length == 0) {
                throw new Error("no guide categories found")
            }

            cachedFull = renderCategories(categories);
            Render(target, cachedFull);
            return categories;
        } catch(err) {
            Render(target, err);
        }
    }

    function SearchResults(target, query, guides) {
        const category = {
            "name": `Results for &ldquo;${query}&rdquo;`,
            "description": "",
            "guides": guides,
            "isSearch": true,
        };

        Render(target, renderCategory(category));
    }

    function SearchError(target, msg) {
        Render(target, `
            <div class="bubble search-error">
                <i class="bi bi-exclamation-circle-fill"></i>
                <div class="bubble-message">${msg}</div>
                <i class="bi bi-exclamation-circle-fill"></i>
            </div>
        `);
    }

    return { Full, Render, SearchError, SearchResults };
})();

document.addEventListener("DOMContentLoaded", e => {
    const target = document.getElementById("results");
    BuildUI.Full(target)
        .then(categories => {
            const search = new Search(categories);
            search.Log();

            const searchField = document.getElementById("search");
            searchField.addEventListener("input", e => {
                const query = e.target.value;
                if (query.length == 0) {
                    BuildUI.Full(target);
                    return;
                }

                if (query.length < MIN_TOKEN_LENGTH) {
                    BuildUI.SearchError(target, `Search term too short (less than ${MIN_TOKEN_LENGTH} characters).`);
                    return;
                }

                const results = search.Find(query);
                if (results.length > 0) {
                    BuildUI.SearchResults(target, query, results);
                } else {
                    BuildUI.SearchError(target, `No results for &ldquo;${query}&rdquo;`);
                }
            })
        });
    
});
