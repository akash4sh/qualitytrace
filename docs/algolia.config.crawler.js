new Crawler({
  rateLimit: 8,
  maxDepth: 10,
  startUrls: ["https://docs.qualitytrace.io/"],
  renderJavaScript: false,
  sitemaps: [],
  ignoreCanonicalTo: false,
  discoveryPatterns: ["https://docs.qualitytrace.io/**"],
  schedule: "at 1:39 PM on Friday",
  actions: [
    {
      indexName: "qualitytrace",
      pathsToMatch: ["https://docs.qualitytrace.io/**"],
      recordExtractor: ({ helpers }) => {
        return helpers.docsearch({
          recordProps: {
            lvl1: ["header h1", "article h1", "main h1", "h1", "head > title"],
            content: ["article p, article li", "main p, main li", "p, li"],
            lvl0: {
              selectors: "",
              defaultValue: "Documentation",
            },
            lvl2: ["article h2", "main h2", "h2"],
            lvl3: ["article h3", "main h3", "h3"],
            lvl4: ["article h4", "main h4", "h4"],
            lvl5: ["article h5", "main h5", "h5"],
            lvl6: ["article h6", "main h6", "h6"],
          },
          aggregateContent: true,
          recordVersion: "v3",
        });
      },
    },
  ],
  initialIndexSettings: {
    qualitytrace: {
      attributesForFaceting: ["type", "lang"],
      attributesToRetrieve: [
        "hierarchy",
        "content",
        "anchor",
        "url",
        "url_without_anchor",
        "type",
      ],
      attributesToHighlight: ["hierarchy", "content"],
      attributesToSnippet: ["content:10"],
      camelCaseAttributes: ["hierarchy", "content"],
      searchableAttributes: [
        "unordered(hierarchy.lvl0)",
        "unordered(hierarchy.lvl1)",
        "unordered(hierarchy.lvl2)",
        "unordered(hierarchy.lvl3)",
        "unordered(hierarchy.lvl4)",
        "unordered(hierarchy.lvl5)",
        "unordered(hierarchy.lvl6)",
        "content",
      ],
      distinct: true,
      attributeForDistinct: "url",
      customRanking: [
        "desc(weight.pageRank)",
        "desc(weight.level)",
        "asc(weight.position)",
      ],
      ranking: [
        "words",
        "filters",
        "typo",
        "attribute",
        "proximity",
        "exact",
        "custom",
      ],
      highlightPreTag: '<span class="algolia-docsearch-suggestion--highlight">',
      highlightPostTag: "</span>",
      minWordSizefor1Typo: 3,
      minWordSizefor2Typos: 7,
      allowTyposOnNumericTokens: false,
      minProximity: 1,
      ignorePlurals: true,
      advancedSyntax: true,
      attributeCriteriaComputedByMinProximity: true,
      removeWordsIfNoResults: "allOptional",
    },
  },
  appId: "L2ILN3GLIL",
  apiKey: "bfcc1af3904668d68edffb4c155de1cb",
});