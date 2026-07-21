package exercisestore

import (
	"fmt"
	"sync"
	"time"
)

type Exercise struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Sets      int       `json:"sets"`
	Reps      int       `json:"reps"`
	Weight    float64   `json:"weight"`
	CreatedAt time.Time `json:"createdat"`
	Tags      []string  `json:"tags"`
}

type ExerciseStore struct {
	sync.Mutex

	Exercises map[int]Exercise
	nextId    int
}

func New() *ExerciseStore {
	es := &ExerciseStore{}
	es.Exercises = make(map[int]Exercise)
	es.nextId = 0
	return es
}

// create an exercise in store
func (es *ExerciseStore) CreateExercise(name string, sets, reps int, weight float64, tags []string) int {
	es.Lock()
	defer es.Unlock()

	e := Exercise{
		ID:        es.nextId,
		Name:      name,
		Sets:      sets,
		Reps:      reps,
		Weight:    weight,
		CreatedAt: time.Now(),
	}
	e.Tags = make([]string, len(tags))
	copy(e.Tags, tags)

	es.Exercises[es.nextId] = e
	es.nextId++
	return e.ID
}

// get exercise from store by id
func (es *ExerciseStore) GetExerciseByID(id int) (Exercise, error) {
	es.Lock()
	defer es.Unlock()

	e, ok := es.Exercises[id]
	if !ok {
		return Exercise{}, fmt.Errorf("Exercise with id: %d was not found", id)
	} else {
		return e, nil
	}
}

// get all exercise from store
func (es *ExerciseStore) GetAllExercises() []Exercise {
	es.Lock()
	defer es.Unlock()

	exerciseList := make([]Exercise, 0, len(es.Exercises))
	for _, v := range es.Exercises {
		exerciseList = append(exerciseList, v)
	}
	return exerciseList
}

// delete exercise by ID
func (es *ExerciseStore) DeleteExercise(id int) error {
	es.Lock()
	defer es.Unlock()

	_, ok := es.Exercises[id]
	if !ok {
		return fmt.Errorf("Exercise with id: %d was not found", id)
	}
	delete(es.Exercises, id)
	return nil
}

// delete all exercises in the store
func (es *ExerciseStore) DeleteAllExercises() int {
	es.Lock()
	defer es.Unlock()
	deleted := len(es.Exercises)
	clear(es.Exercises)
	return deleted
}

// gets all exercises that have the given tag
func (es *ExerciseStore) GetExerciseByTag(tag string) []Exercise {
	es.Lock()
	defer es.Unlock()

	var res []Exercise

	for _, e := range es.Exercises {
		for _, t := range e.Tags {
			if t == tag {
				res = append(res, e)
			}
		}
	}
	return res
}

// gets all exercises that have the given created at date
func (es *ExerciseStore) GetExerciseByDate(year int, month time.Month, day int) []Exercise {
	es.Lock()
	defer es.Unlock()

	var res []Exercise

	for _, e := range es.Exercises {
		y, m, d := e.CreatedAt.Date()
		if y == year && m == month && d == day {
			res = append(res, e)
		}
	}

	return res
}
