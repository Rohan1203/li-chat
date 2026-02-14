package db

func (r *Repository) CreateUser(username, passwordHash string) error {
	_, err := r.db.Exec(
		"INSERT INTO users(username, password_hash) VALUES ($1, $2)",
		username, passwordHash,
	)
	return err
}

func (r *Repository) GetUserForLogin(username string) (int64, string, error) {
	var id int64
	var hash string

	err := r.db.QueryRow(
		"SELECT id, password_hash FROM users WHERE username = $1",
		username,
	).Scan(&id, &hash)

	return id, hash, err
}

func (r *Repository) GetUserByID(userID int64) (string, error) {
	var username string

	err := r.db.QueryRow(
		"SELECT username FROM users WHERE id = $1",
		userID,
	).Scan(&username)

	return username, err
}
