package db

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/MahdiiTaheri/classnama-backend/internal/store"
)

// Sample data for seeding
var (
	firstNames = []string{"John", "Alice", "Bob", "Emma", "Liam", "Sophia", "David", "Olivia"}
	lastNames  = []string{"Doe", "Smith", "Johnson", "Brown", "Williams", "Jones", "Garcia"}
	subjects   = []string{"Math", "Physics", "Chemistry", "Biology", "History", "English"}
	classes    = []string{"1A", "1B", "2A", "2B", "3A", "3B"}
	roles      = []string{"admin", "manager"}
)

// Seed populates the database
func Seed(store store.Storage) {
	ctx := context.Background()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 1. Seed Execs
	execs := generateExecs(15, rng)
	for _, e := range execs {
		if err := e.Password.Set("password123"); err != nil {
			log.Println("Error setting exec password:", err)
			continue
		}
		if err := store.Execs.Create(ctx, e); err != nil {
			log.Println("Error creating exec:", err)
		}
	}

	// 2. Seed Teachers
	teachers := generateTeachers(30, rng)
	for _, t := range teachers {
		if err := t.Password.Set("password123"); err != nil {
			log.Println("Error setting teacher password:", err)
			continue
		}
		if err := store.Teachers.Create(ctx, t); err != nil {
			log.Println("Error creating teacher:", err)
		}
	}

	// 3. Seed Students
	students := generateStudents(300, teachers, rng)
	for _, s := range students {
		if err := s.Password.Set("password123"); err != nil {
			log.Println("Error setting student password:", err)
			continue
		}
		if err := store.Students.Create(ctx, s); err != nil {
			log.Println("Error creating student:", err)
		}
	}

	log.Println("Seeding complete!")
}

// Generate random exec users
func generateExecs(n int, rng *rand.Rand) []*store.Exec {
	execs := make([]*store.Exec, n)
	for i := range n {
		execs[i] = &store.Exec{
			FirstName: firstNames[rng.Intn(len(firstNames))],
			LastName:  lastNames[rng.Intn(len(lastNames))],
			Email:     fmt.Sprintf("exec%d@example.com", i),
			Role:      store.Role(roles[rng.Intn(len(roles))]),
		}
	}
	return execs
}

// Generate random teachers
func generateTeachers(n int, rng *rand.Rand) []*store.Teacher {
	teachers := make([]*store.Teacher, n)
	for i := range n {
		teachers[i] = &store.Teacher{
			FirstName:   firstNames[rng.Intn(len(firstNames))],
			LastName:    lastNames[rng.Intn(len(lastNames))],
			Email:       fmt.Sprintf("teacher%d@example.com", i),
			Subject:     subjects[rng.Intn(len(subjects))],
			PhoneNumber: fmt.Sprintf("+12345678%02d", i),
			HireDate:    time.Now().AddDate(-rng.Intn(5), 0, 0), // random hire date in last 5 years
		}
	}
	return teachers
}

// Generate random students assigned to teachers
func generateStudents(n int, teachers []*store.Teacher, rng *rand.Rand) []*store.Student {
	students := make([]*store.Student, n)
	for i := range n {
		teacher := teachers[rng.Intn(len(teachers))]
		students[i] = &store.Student{
			FirstName:         firstNames[rng.Intn(len(firstNames))],
			LastName:          lastNames[rng.Intn(len(lastNames))],
			Email:             fmt.Sprintf("student%d@example.com", i),
			Class:             classes[rng.Intn(len(classes))],
			BirthDate:         time.Now().AddDate(-10-rng.Intn(8), 0, 0), // age 10–18
			Address:           fmt.Sprintf("Street %d", i),
			ParentName:        firstNames[rng.Intn(len(firstNames))] + " " + lastNames[rng.Intn(len(lastNames))],
			ParentPhoneNumber: fmt.Sprintf("+98765432%02d", i),
			PhoneNumber:       func() *string { s := fmt.Sprintf("+98765432%02d", i); return &s }(),
			TeacherID:         teacher.ID,
		}
	}
	return students
}
