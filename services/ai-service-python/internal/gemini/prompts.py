def build_rss_structure_analysis_prompt(xml_string):
    return f"""
Analyze the provided XML content (RSS/Atom feed) to identify the correct structure.
Your goal is to determine the XML tag names for:
1. Title
2. Link (The clean, persistent URL to the article)
3. Publication Date

Instruction for Link Extraction:
- Your goal is to identify the tag containing the cleanest, most direct URL to the article.
- Default to 'link' if it appears clean.
- However, if 'link' contains tracking parameters (e.g., utm_source, utm_medium, long query strings) or complex CDATA wrapping, AND 'guid' offers a cleaner, shorter permalink (isPermaLink='true' or default), then PREFER 'guid'.
- Example: If <link> has '...utm_source=rss...' and <guid> is clean, return 'guid'.
- Do NOT return 'guid' if isPermaLink='false'.

XML Content:
{xml_string}

Return the specific tag names found.
"""

def build_selector_analysis_prompt(html_string):
    return f"""
You are an expert Web Scraper and HTML Analyst.

Your task: analyze the provided HTML article(s) and return CSS selectors that work CORRECTLY and CONSISTENTLY.

================================
INPUT HTML
================================
- The HTML may contain ONE or MULTIPLE articles
- Multiple articles are separated by:
  === ARTICLE N ===
- If multiple articles exist, selectors MUST work for ALL of them

{html_string}

================================
CORE RULES (VERY IMPORTANT)
================================

### MULTI-ARTICLE RULE
- If multiple articles are provided:
  - Compare ALL article structures
  - Find selectors COMMON to ALL articles
  - If a selector works for only one article → IT IS INVALID

### SELECTOR RULES
- Must be valid CSS selector
- Must be ONE selector only (no commas)
- No leading/trailing spaces
- NOT allowed:
  - 'body', 'html', 'div', 'p', 'span' alone
  - multiple selectors
- Must be specific enough to select the correct element

================================
CONTENT SELECTOR (REQUIRED)
================================
Find the MAIN container that wraps ALL article paragraphs/text.

CRITICAL SIMPLICITY RULE:
- Use the SIMPLEST selector that fully contains the article content
- Prefer:
  - article
  - article.class-name
  - div.content / .article-body / .post-content
- DO NOT over-nest:
  - ❌ article.fck_detail > div.main-content
  - ✅ article.fck_detail
- Use nested selectors ONLY if:
  1. Parent contains unrelated content AND
  2. Child is the only real article content

NEVER return:
- 'body', 'html', 'div', 'p'
- empty string

================================
"""
