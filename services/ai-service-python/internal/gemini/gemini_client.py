import google.generativeai as genai
import logging

class GeminiClient:
    def __init__(self, api_key, model_name):
        if not model_name:
            model_name = "gemini-1.5-flash"
        
        genai.configure(api_key=api_key)
        self.model = genai.GenerativeModel(model_name)

    def generate_content(self, prompt, config=None):
        try:
            response = self.model.generate_content(prompt, generation_config=config)
            if not response.candidates:
                raise Exception("no candidates returned")
            return response.text
        except Exception as e:
            logging.error(f"failed to generate content: {e}")
            raise e
