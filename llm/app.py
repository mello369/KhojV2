from flask import Flask, request, jsonify
import argparse
import os
import replicate
import concurrent.futures
import re
app = Flask(__name__)

# Function to generate LLM response for the first prompt
def generate_llm_response(prompt):
    return replicate.run(
        'a16z-infra/llama13b-v2-chat:df7690f1994d94e96ad9d568eac121aecf50684a0b0963b25a41cc40061269e5',
        input={
            "prompt": prompt,
            "temperature": 0.1,
            "top_p": 0.9,
            "max_length": 300,
            "repetition_penalty": 1
        }
    )

def remove_circular_brackets(text):
    return re.sub(r'\([^)]*\)','',text)

@app.route('/hello')
def hello():
    return "hello"

@app.route('/generate_llama_responses',methods=['POST'])
def generate_responses():
    try:
        # Get the keyword from the request JSON data
        #keyword = request.json['keyword']
        return [{'keywords': ['split red lentils ', 'yellow split peas ', 'rice', 'ghee or vegetable oil', 'cumin seeds', 'coriander seeds', 'cardamom', 'cinnamon', 'cloves', 'bay leaves', 'star anise', 'fennel seeds', 'garam masala', 'turmeric', 'red chili powder', 'garlic', 'ginger', 'onion', 'tomato', 'tamarind paste', 'lemon juice', 'salt', 'water'], 'name': 'dalmakhani'}]
        os.environ["REPLICATE_API_TOKEN"] = "r8_cTLmg7ffonyo98U5hKc53OSw73Wxz3n1GxqFF"
        keyword = request.args.get('keyword')
        data = request.get_json()
        max = data[0]["Confidence"]
        max_class = data[0]["Prediction"]

        for obj in data :
            if obj["Confidence"] > max :
                max = obj["Confidence"]
                max_class = obj["Prediction"]

        response =[]
        prompt_input = f"Ingredients to purchase for preparing {max_class}  without mentioning quantity ?"
        with concurrent.futures.ThreadPoolExecutor(max_workers=3) as executor:
        # Execute all three LLM responses concurrently
            future1 = executor.submit(generate_llm_response, f" {prompt_input}")

            full_response = future1.result()

        ingredient_pattern = r'\* (.*?)(?:\n|\*)'


        full_text = "".join(full_response)

        ingredients_matches = re.findall(ingredient_pattern, full_text, re.IGNORECASE)
        
        ingredients_list = [ingredient.strip().lower() for ingredient in ingredients_matches]

        response_data = {
                "name": max_class,
                "keywords": ingredients_list,
                #"Keywords": keyword_matches,
                #"third_response": third_text
            }
        
        for key,value in response_data.items():
            if isinstance(value, list):
                response_data[key] = [remove_circular_brackets(item) for item in value]
            elif isinstance(value,str):
                response_data[key] = remove_circular_brackets(value)
        
        response.append(response_data)
          

        for obj in data :
            if obj["Prediction"] == max_class :
                continue
            prompt_input = f"Ingredients to purchase for preparing {obj['Prediction']}  without mentioning quantity ?"
            with concurrent.futures.ThreadPoolExecutor(max_workers=3) as executor:
            # Execute all three LLM responses concurrently
                future1 = executor.submit(generate_llm_response, f" {prompt_input}")

                full_response = future1.result()

            ingredient_pattern = r'\* (.*?)(?:\n|\*)'


            full_text = "".join(full_response)

            ingredients_matches = re.findall(ingredient_pattern, full_text, re.IGNORECASE)
        
            ingredients_list = [ingredient.strip().lower() for ingredient in ingredients_matches]

            response_data = {
                "name": obj["Prediction"],
                "keywords": ingredients_list,
                #"Keywords": keyword_matches,
                #"third_response": third_text
            }
            for key,value in response_data.items():
                if isinstance(value, list):
                    response_data[key] = [remove_circular_brackets(item) for item in value]
                elif isinstance(value,str):
                    response_data[key] = remove_circular_brackets(value)

            response.append(response_data)
              

        

        # Prompts
        # pre_prompt = ""
        
        #second_prompt = f"Generate relevant keywords related to {keyword}  for quick commerce"
        

        # Set the REPLICATE_API_TOKEN environment variable
        

        # Create a ThreadPoolExecutor with 3 threads



        

        # Create a response dictionary

        print(response)
        



        return jsonify(response)
    except Exception as e:
        return jsonify({"error": str(e)})

if __name__ == '__main__':
    app.run(debug=True,port=4000,host='0.0.0.0')
    parser = argparse.ArgumentParser(description="Flask api exposing yolov5 model")
    parser.add_argument("--port", default=3000, type=int, help="port number")
    args = parser.parse_args()

