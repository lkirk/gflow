package main

// type Dependency struct {
// 	InJobs  []*Job            `json:"in_jobs"`
// 	OutJobs []*Job            `json:"out_jobs"`
// 	Files   []*FileDependency `json:"files"`
// }

// type FileDependency struct {
// 	Name         string `json:"name"`
// 	AssertExists bool   `json:"assert_exists"`
// }

// func OneToMany(ij *Job, oj []*Job, f []*FileDependency) *Dependency {
// 	ijs := make([]*Job, 0)
// 	if ij != nil {
// 		ijs = append(ijs, ij)
// 	}
// 	return &Dependency{ijs, oj, f}
// }

// func ManyToOne(ij, oj []*Job, f []*FileDependency) *Dependency {
// 	return &Dependency{ij, oj, f}
// }

// func OneToOne(ij, oj *Job, f []*FileDependency) *Dependency {
// 	ijs := make([]*Job, 0)
// 	if ij != nil {
// 		ijs = append(ijs, ij)
// 	}
// 	ojs := make([]*Job, 0)
// 	if oj != nil {
// 		ojs = append(ojs, oj)
// 	}
// 	return &Dependency{ijs, ojs, f}
// }
