-- Initialize the database with some sample tables and data

CREATE DATABASE IF NOT EXISTS otel_example;
USE otel_example;

-- Sample users table
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    bio TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Insert some sample data
INSERT INTO users (name, email, bio) VALUES 
    ('John Doe', 'john@example.com', 'I am a software engineer'),
    ('Jane Smith', 'jane@example.com', 'I am a salesperson'),
    ('Bob Johnson', 'bob@example.com', 'I am a manager');

