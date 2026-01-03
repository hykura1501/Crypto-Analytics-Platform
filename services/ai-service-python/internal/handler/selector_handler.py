import logging
import json
import os
import time
from internal.db.database import db
from internal.gemini.prompts import build_selector_analysis_prompt

class SelectorHandler:
    def __init__(self, gemini_client):
        self.gemini_client = gemini_client

    def handle(self, msg):
        source_id = msg.get("source_id")
        html_string = msg.get("html_string")

        prompt = build_selector_analysis_prompt(html_string)

        # Write prompt to file for debugging
        prompt_dir = "/tmp/ai-prompts/selector"
        os.makedirs(prompt_dir, exist_ok=True)
        timestamp = time.strftime("%Y%m%d-%H%M%S")
        prompt_file = os.path.join(prompt_dir, f"{source_id}_{timestamp}.txt")
        try:
            with open(prompt_file, "w") as f:
                f.write(prompt)
            logging.info(f"📝 Prompt written to: {prompt_file}")
        except Exception as e:
            logging.warning(f"⚠️  Failed to write prompt to file: {e}")

        selector_response_schema = {
            "type": "object",
            "properties": {
                "summary_selector": {
                    "type": "string",
                    "description": "Single CSS selector for article summary/excerpt. Must be valid CSS selector (e.g., '.excerpt', 'p.lead'). Return empty string \"\" if not found. Must NOT be generic tags like 'body', 'div', 'p' without classes.",
                },
                "content_selector": {
                    "type": "string",
                    "description": "REQUIRED: Single CSS selector for main article content container. Must be specific (e.g., 'article', '.article-body', '#content-main'). Must NOT be generic tags like 'body', 'html', 'div', 'p' without classes/ids.",
                },
                "author_selector": {
                    "type": "string",
                    "description": "Single CSS selector for author name (e.g., '.author', 'span.byline'). Return empty string \"\" if not found. Must NOT be generic tags.",
                },
                "tags_selector": {
                    "type": "string",
                    "description": "Single CSS selector for tags/categories container (e.g., '.tags', 'ul.tag-list'). Return empty string \"\" if not found. Must NOT be generic tags.",
                },
            },
            "required": ["content_selector"],
        }

        try:
            content = self.gemini_client.generate_content(
                prompt,
                config={"response_mime_type": "application/json", "response_schema": selector_response_schema}
            )

            response = json.loads(content)
            logging.info(f"CSS selector analysis result: {response}")

            # Update database
            with db.conn.cursor() as cur:
                query = "UPDATE sources SET summary_selector = %s, content_selector = %s, author_selector = %s, tags_selector = %s WHERE source_id = %s"
                cur.execute(query, (
                    response.get("summary_selector", ""),
                    response.get("content_selector", ""),
                    response.get("author_selector", ""),
                    response.get("tags_selector", ""),
                    source_id
                ))
                db.conn.commit()

            logging.info(f"✅ Updated CSS selectors for source {source_id}")
            return response

        except Exception as e:
            logging.error(f"Error handling selector analysis: {e}")
            db.conn.rollback()
            return None
