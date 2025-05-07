# Use a lightweight Python image
FROM python:3.9-slim

# Set the working directory
WORKDIR /app

# Copy the Flask app code into the container
COPY ratings.py /app

# Install Flask
RUN pip install flask

# Expose the port the app will run on
EXPOSE 9080

# Run the Flask app
CMD ["python", "ratings.py"]
