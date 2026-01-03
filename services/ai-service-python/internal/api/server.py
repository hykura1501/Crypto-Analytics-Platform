from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from config import config
from internal.handler.rss_handler import RssHandler
from internal.handler.selector_handler import SelectorHandler

class RssAnalysisRequest(BaseModel):
    xml_string: str

class SelectorAnalysisRequest(BaseModel):
    html_string: str

class Server:
    def __init__(self, rss_handler: RssHandler, selector_handler: SelectorHandler):
        self.app = FastAPI()
        self.rss_handler = rss_handler
        self.selector_handler = selector_handler
        self.setup_routes()

    def setup_routes(self):
        @self.app.get("/api/v1/ai/health")
        async def health():
            return {"service": "ai-service", "status": "running"}

        @self.app.post("/api/v1/ai/rss/{source_id}")
        async def handle_rss(source_id: str, req: RssAnalysisRequest):
            msg = {
                "source_id": source_id,
                "xml_string": req.xml_string
            }
            response = self.rss_handler.handle(msg)
            if response is None:
                raise HTTPException(status_code=500, detail="RSS analysis failed")
            
            return {
                "message": "RSS analysis completed",
                "response": response
            }

        @self.app.post("/api/v1/ai/selector/{source_id}")
        async def handle_selector(source_id: str, req: SelectorAnalysisRequest):
            msg = {
                "source_id": source_id,
                "html_string": req.html_string
            }
            response = self.selector_handler.handle(msg)
            if response is None:
                raise HTTPException(status_code=500, detail="CSS selector analysis failed")
            
            return {
                "message": "CSS selector analysis completed",
                "response": response
            }

    def run(self):
        import uvicorn
        uvicorn.run(self.app, host="0.0.0.0", port=config.api.port)
