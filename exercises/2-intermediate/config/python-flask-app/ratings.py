# ratings.py
from flask import Flask, jsonify
app = Flask(__name__)

@app.route('/ratings/<int:id>')
def get_ratings(id):
    return jsonify({
        "id": id,
        "ratings": {
            "Reviewer1": 5,
            "Reviewer2": 4
        },
        "version": "v2-modified"
    })

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=9080)
