const MIN_TOKEN_LENGTH = 3;
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
                    .concat(" ", cat.name, " ", cat.description)
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
        query = query.toLowerCase();
        if (!(query in this.index)) {
            console.log(`query ${query} not in index`);
            return [];
        }

        const indices = this.index[query];
        const results = [];
        for (let i = 0; i < indices.length; i++) {
            results.push(this.guides[indices[i]]);
        }
        console.log(`results for ${query}:`, indices, results);

        return results;
    }
}

const BuildUI = (() => {
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
                return `<li class="attachment-item"><i class="bi bi-card-image"></i> <a href="responses/${item.filename}" target="_BLANK">${item.filename}</a> <span class="att-type">(${item.contenttype})</span></li>`;
            case "markdown":
                return `<li class="attachment-item"><i class="bi bi-file-earmark-text-fill"></i> <a href="responses/${item.filename}" target="_BLANK">${item.filename}</a> <span class="att-type">(markdown/text)</span></li>`;
            case "link":
                return `<li class="attachment-item"><i class="bi bi-box-arrow-up-right"></i> <a href="${item.link}" target="_BLANK">${linkText(item.link)}</a> <span class="att-type">(external link)</span></li>`;
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
            "name": `Results for ${query}`,
            "description": "",
            "guides": guides,
        };

        Render(target, renderCategory(category));
    }

    return { Full, Render, SearchResults };
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
                    BuildUI.Render(target, `<p>Search term too short (less than ${MIN_TOKEN_LENGTH} characters).</p>`); 
                    return;
                }

                const results = search.Find(query);
                if (results.length > 0) {
                    BuildUI.SearchResults(target, query, results);
                } else {
                    BuildUI.Render(target, `<p>No results found.</p>`)
                }
            })
        });
    
});
