package auth

// HashPassword создает хеш пароля используя bcrypt
func HashPassword(password string) (string, error) {
	// TODO: Реализовать хеширование пароля
	// 1. Использовать bcrypt.GenerateFromPassword с cost = 10
	// 2. Вернуть хеш пароля в виде строки
	// 3. Если ошибка - вернуть ошибку
	return "", nil
}

// VerifyPassword проверяет что пароль соответствует хешу
func VerifyPassword(hashedPassword, password string) bool {
	// TODO: Реализовать проверку пароля
	// 1. Использовать bcrypt.CompareHashAndPassword
	// 2. Вернуть true если пароль совпадает, false иначе
	return false
}
