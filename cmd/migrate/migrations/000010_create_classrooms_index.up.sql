CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS uniq_classrooms_grade_name
    ON classrooms(grade, name);