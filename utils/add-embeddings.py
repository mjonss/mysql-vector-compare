import mysql.connector
import requests
import json
from typing import List, Tuple

# Configuration for database connection
DB_CONFIG = {
    "host": "127.0.0.1",
    "user": "root",
    "port": 4000,
    "password": "",
    "database": "test"
}

# Ollama server configuration (make sure Ollama is running locally or update the URL accordingly)
OLLAMA_SERVER_URL = "http://localhost:11434/api/embed"

def fetch_markdown_content() -> List[Tuple[str, str]]:
    """
    Fetches all rows from the markdown_files table.
    Returns a list of tuples containing (filename, content).
    """
    try:
        conn = mysql.connector.connect(**DB_CONFIG)
        cursor = conn.cursor()
        
        query = "SELECT id, content FROM markdown_files"
        cursor.execute(query)
        
        results = cursor.fetchall()
        print(f"Retrieved {len(results)} rows from the database.")
        return results
        
    except Exception as e:
        print(f"Error fetching data from database: {e}")
        return []
    finally:
        if 'conn' in locals():
            conn.close()

def generate_embeddings(text: str) -> List[float]:
    """
    Generates vector embeddings for a given text using the Ollama model.
    Returns a list of float values representing the embedding.
    """
    payload = {
        #"model": "snowflake-arctic-embed2",
        "model": "nomic-embed-text",
        "input": f"{text}",
    }

    try:
        response = requests.post(OLLAMA_SERVER_URL, json=payload)
        response.raise_for_status()
        
        result = json.loads(response.text)
        if 'embeddings' in result:
            return result['embeddings'][0]
        else:
            print(f"No embedding found in the response: {response.text}")
            return []
            
    except Exception as e:
        print(f"Error generating embeddings: {e}")
        return []

def store_embeddings(row_id: str, embedding: List[float]) -> None:
    """
    Stores the filename, content, and its embedding vector back into the database.
    Creates a new table `markdown_embeddings` if it doesn't exist.
    """
    try:
        conn = mysql.connector.connect(**DB_CONFIG)
        cursor = conn.cursor()
        
        emb_string = "[" + ",".join(map(str, embedding)) + "]"
        print(f"row_id: {row_id}")
        print(f"emb_string: {emb_string}")
        update_query = f"UPDATE markdown_files SET `nomic-embed-text` = VEC_FROM_TEXT('{emb_string}') WHERE id = {row_id}"
        print(f"query: {update_query}")
        cursor.execute(update_query)
        #cursor.execute(update_query, (emb_string, row_id))
        
        conn.commit()
        print(f"Stored embeddings for {row_id}")
        
    except Exception as e:
        print(f"Error storing embeddings: {e}")
    finally:
        if 'conn' in locals():
            conn.close()

def main() -> None:
    """
    Main function that processes all markdown files, generates embeddings, and stores them.
    """
    # Fetch all markdown content
    markdown_data = fetch_markdown_content()
    
    for row_id, content in markdown_data:
        print(f"Processing {row_id}")
        
        # Generate embeddings
        embedding = generate_embeddings(content)
        if not embedding:
            continue
            
        # Store embeddings back into the database
        store_embeddings(row_id, embedding)

if __name__ == "__main__":
    main()

