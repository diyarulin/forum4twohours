# Forum Project

## Objectives
This project involves creating a web forum that allows:
- Communication between users.
- Associating categories with posts.
- Liking and disliking posts and comments.
- Filtering posts.

## Technologies
- **Go** for backend development.
- **SQLite** for database management.
- **Docker** for containerization.
- **HTML, CSS, JavaScript** for frontend (without frameworks).
- **UUID, bcrypt, sqlite3** as allowed Go packages.

## Features
### Authentication
- User registration requires an email, username, and password.
- If an email is already taken, an error is returned.
- Passwords should be encrypted when stored (**Bonus**).
- Login session is managed using cookies with an expiration date.

### Communication
- Only registered users can create posts and comments.
- Posts can have one or more categories.
- All users (registered or not) can view posts and comments.

### Likes & Dislikes
- Only registered users can like or dislike posts and comments.
- The number of likes and dislikes is visible to all users.

### Filtering
Users can filter posts by:
- Categories.
- Posts created by them (**Registered users only**).
- Liked posts (**Registered users only**).

## Requirements
- **Database:** Must use SQLite with at least one `SELECT`, `CREATE`, and `INSERT` query.
- **Error Handling:** Handle website errors, HTTP status codes, and technical issues.
- **Security:** Implement proper authentication and data encryption where applicable.
- **Best Practices:** Code should follow good programming practices and include unit tests.
- **Frontend:** No frontend frameworks like React, Angular, or Vue are allowed.

## Setup & Installation
1. Clone the repository:
   ```sh
   git clone git@github.com:diyarulin/forum4twohours.git
   git switch check_merge
   cd forum4twohours
   ```
2. Run without Docker:
   ```sh
   go run ./cmd/web
   ```
3. Build and run with Docker:
   ```sh
   sudo docker-compose build .
   sudo docker-compose up
   ```
4. Access the forum at `http://localhost:8080`

