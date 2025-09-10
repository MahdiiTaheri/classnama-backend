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
	firstNames = []string{
		"John", "Alice", "Bob", "Emma", "Liam", "Sophia", "David", "Olivia",
		"Michael", "Isabella", "Daniel", "Mia", "James", "Charlotte", "William", "Amelia",
		"Henry", "Evelyn", "Alexander", "Harper", "Matthew", "Abigail", "Joseph", "Ella",
		"Samuel", "Avery", "Owen", "Scarlet", "Lucas", "Victoria",
	}

	lastNames = []string{
		"Doe", "Smith", "Johnson", "Brown", "Williams", "Jones", "Garcia",
		"Miller", "Davis", "Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez",
		"Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson",
	}

	subjects = []string{
		"Math", "Physics", "Chemistry", "Biology", "History", "English",
		"Geography", "Computer Science", "Art", "Music", "Physical Education",
		"Economics", "Philosophy", "Psychology", "Sociology", "Literature", "French", "Spanish",
	}

	roles = []string{"admin", "manager"}

	classroomNames = []string{
		"1A", "1B", "2A", "2B", "3A", "3B", "4A", "4B", "5A", "5B", "6A", "6B",
	}
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

	// 2. Seed Teachers (one per classroom)
	teachers := generateTeachers(10, rng)
	for _, t := range teachers {
		if err := t.Password.Set("password123"); err != nil {
			log.Println("Error setting teacher password:", err)
			continue
		}
		if err := store.Teachers.Create(ctx, t); err != nil {
			log.Println("Error creating teacher:", err)
		}
	}

	// 3. Seed Classrooms with assigned TeacherID
	classrooms := generateClassroomsWithTeachers(teachers, rng)
	for _, c := range classrooms {
		if err := store.Classrooms.Create(ctx, c); err != nil {
			log.Println("Error creating classroom:", err)
		}
	}

	// 4. Seed Students
	students := generateStudents(300, classrooms, rng)
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
	for i := 0; i < n; i++ {
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
	for i := 0; i < n; i++ {
		teachers[i] = &store.Teacher{
			FirstName:   firstNames[rng.Intn(len(firstNames))],
			LastName:    lastNames[rng.Intn(len(lastNames))],
			Email:       fmt.Sprintf("teacher%d@example.com", i),
			Subject:     subjects[rng.Intn(len(subjects))],
			PhoneNumber: fmt.Sprintf("+12345678%02d", i),
			HireDate:    time.Now().AddDate(-rng.Intn(5), 0, 0),
		}
	}
	return teachers
}

// Generate classrooms with one teacher each
func generateClassroomsWithTeachers(teachers []*store.Teacher, rng *rand.Rand) []*store.Classroom {
	classrooms := make([]*store.Classroom, len(teachers))
	for i, t := range teachers {
		classrooms[i] = &store.Classroom{
			Name:      classroomNames[rng.Intn(len(classroomNames))],
			Capacity:  int64(20 + rng.Intn(10)),
			Grade:     int64(rng.Intn(12) + 1),
			TeacherID: t.ID, // assign teacher
		}
	}
	return classrooms
}

// Generate students assigned to classrooms
func generateStudents(n int, classrooms []*store.Classroom, rng *rand.Rand) []*store.Student {
	students := make([]*store.Student, n)
	for i := 0; i < n; i++ {
		classroom := classrooms[rng.Intn(len(classrooms))]
		students[i] = &store.Student{
			FirstName:         firstNames[rng.Intn(len(firstNames))],
			LastName:          lastNames[rng.Intn(len(lastNames))],
			Email:             fmt.Sprintf("student%d@example.com", i),
			ClassRoomID:       classroom.ID,
			BirthDate:         time.Now().AddDate(-10-rng.Intn(8), 0, 0),
			Address:           fmt.Sprintf("Street %d", i),
			ParentName:        firstNames[rng.Intn(len(firstNames))] + " " + lastNames[rng.Intn(len(lastNames))],
			ParentPhoneNumber: fmt.Sprintf("+98765432%02d", i),
			PhoneNumber:       func() *string { s := fmt.Sprintf("+98765432%02d", i); return &s }(),
			TeacherID:         classroom.TeacherID, // follow classroom
		}
	}
	return students
}
