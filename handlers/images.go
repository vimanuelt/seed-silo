package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ngageoint/seed-silo/database"
	"github.com/ngageoint/seed-silo/models"
)

func ListImages(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	imageList := []models.SimpleImage{}
	images := models.ReadSimpleImages(db)
	imageList = append(imageList, images...)

	respondWithJSON(w, http.StatusOK, imageList)
}

type RankedResult struct {
	Score int
	Image models.Image
}

type ByScore []RankedResult

func (s ByScore) Len() int {
	return len(s)
}
func (s ByScore) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByScore) Less(i, j int) bool {
	return s[i].Score > s[j].Score
}

func SearchImages(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	vars := mux.Vars(r)
	query := vars["query"]

	if query == "" || query == "+" {
		ListImages(w, r)
		return
	}

	terms := strings.Split(query, "+")

	images := models.ReadImages(db)

	rankedResults := []RankedResult{}
	for _, img := range images {
		score := 0
		for _, term := range terms {
			if strings.Contains(img.FullName, term) {
				score += 10
			}
			if strings.Contains(img.Org, term) {
				score += 10
			}
			seed := img.Seed

			if strings.Contains(fmt.Sprintf("%v", img.Seed), term) {
				score += 1
			}

			if strings.Contains(seed.Job.Name, term) {
				score += 10
			}
			if strings.Contains(seed.Job.Title, term) {
				score += 5
			}

			if strings.Contains(seed.Job.Description, term) {
				score += 5
			}

			if strings.Contains(fmt.Sprintf("%s", seed.Job.Tags), term) {
				score += 10
			}

			if strings.Contains(fmt.Sprintf("%s", seed.Job.Maintainer), term) {
				score += 5
			}

		}
		if score > 0 {
			rankedResults = append(rankedResults, RankedResult{Score: score, Image: img})
		}
	}

	sort.Sort(ByScore(rankedResults))

	results := []models.SimpleImage{}
	for _, res := range rankedResults {
		simple := models.SimplifyImage(res.Image)
		results = append(results, simple)
	}

	respondWithJSON(w, http.StatusOK, results)
}

func Image(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	img, err := models.ReadImage(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "No image found with that ID")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	respondWithJSON(w, http.StatusOK, img)
}

func ImageManifest(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	image, err := models.ReadImage(db, id)

	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "No image found with that ID")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	respondWithJSON(w, http.StatusOK, image.Seed)
}
