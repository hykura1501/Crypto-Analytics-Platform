import logging
import json
from internal.db.database import db
from internal.gemini.prompts import build_rss_structure_analysis_prompt

class RssHandler:
    def __init__(self, gemini_client):
        self.gemini_client = gemini_client

    def handle(self, msg):
        # msg is already a dict/object from the main loop
        source_id = msg.get("source_id")
        xml_string = msg.get("xml_string")

        prompt = build_rss_structure_analysis_prompt(xml_string)

        rss_response_schema = {
            "type": "object",
            "properties": {
                "title_tag": {
                    "type": "string",
                    "description": "The specific XML tag name used for the title of an item (e.g. 'title'). Return only the tag name.",
                },
                "link_tag": {
                    "type": "string",
                    "description": "The specific XML tag name used for the URL/permalink of an item. Prefer 'link'. If 'guid' is explicitly a better permalink (isPermaLink='true') and 'link' is missing, use 'guid'. Return only the tag name.",
                },
                "pub_date_tag": {
                    "type": "string",
                    "description": "The specific XML tag name used for the publication date of an item (e.g. 'pubDate', 'dc:date'). Return only the tag name.",
                },
            },
            "required": ["title_tag", "link_tag", "pub_date_tag"],
        }

        try:
            content = self.gemini_client.generate_content(
                prompt, 
                config={"response_mime_type": "application/json", "response_schema": rss_response_schema}
            )
            
            response = json.loads(content)
            logging.info(f"RSS structure analysis result: {response}")

            # Update database
            with db.conn.cursor() as cur:
                query = "UPDATE sources SET title_tag = %s, link_tag = %s, pub_date_tag = %s WHERE source_id = %s"
                cur.execute(query, (response["title_tag"], response["link_tag"], response["pub_date_tag"], source_id))
                db.conn.commit()
            
            logging.info(f"✅ Updated RSS structure for source {source_id}")
            return response

        except Exception as e:
            logging.error(f"Error handling RSS analysis: {e}")
            db.conn.rollback()
            return None
