package documentstore

import (
	"context"
	"errors"
	"fmt"
	"gitlab.com/shitposting/autoposting-bot/documentstore/entities"
	fpcompare "gitlab.com/shitposting/fingerprinting/comparer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/xerrors"
	"math"
	"time"
)

func AddPost(addedBy int32, media entities.Media, caption string, collection *mongo.Collection) error {

	post := entities.Post{
		AddedBy: addedBy,
		Media:   media,
		Caption: caption,
		AddedAt: time.Now(),
	}

	//
	ctx, cancelCtx := context.WithTimeout(context.Background(), opDeadline)
	defer cancelCtx()

	_, err := collection.InsertOne(ctx, post)
	if err != nil {
		err = fmt.Errorf("AddPost: %v", err)
	}

	return err

}

//// UpdatePostCaptionByFileID updates the caption of a post given its fileID
//func UpdatePostCaptionByFileID(fileID, caption string) bool {
//
//}
//

// FindPostByFeatures finds a post by its features
func FindPostByFeatures(histogram []float64, pHash string, approximation float64, collection *mongo.Collection) (post entities.Post, err error) {

	//
	if histogram == nil {
		err = xerrors.New("FindPostByFeatures: histogram was nil")
		return
	}

	if pHash == "" {
		err = xerrors.New("FindPostByFeatures: pHash was empty")
		return
	}

	//
	average, sum := entities.GetHistogramAverageAndSum(histogram)
	minAvg := math.Trunc(average - 1)
	maxAvg := math.Ceil(average + 1)
	minSum := math.Trunc(sum - (sum * approximation))
	maxSum := math.Ceil(sum + (sum * approximation))

	//
	filter := bson.D{
		{
			Key: "media.histogramaverage",
			Value: bson.D{
				{"$gte", minAvg},
				{"$lte", maxAvg},
			},
		},
		{Key: "media.histogramsum",
			Value: bson.D{
				{"$gte", minSum},
				{"$lte", maxSum},
			},
		},
	}

	//
	ctx, cancelCtx := context.WithTimeout(context.Background(), opDeadline)
	defer cancelCtx()

	//TODO: ordinare secondo qualcosa i dati

	//
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		err = xerrors.Errorf("FindPostByFeatures: unable to retrieve post: %s", err)
		return
	}

	post, err = findBestMatch(pHash, cursor)
	if err != nil {
		err = xerrors.Errorf("FindMediaByFeatures: %s", err)
		return
	}

	return

}


// FindPostByFileID retrieves a post via its fileID
func FindPostByUniqueID(uniqueID string, collection *mongo.Collection) (post entities.Post, err error) {

	if uniqueID == "" {
		return post, errors.New("uniqueID empty")
	}

	//
	ctx, cancelCtx := context.WithTimeout(context.Background(), opDeadline)
	defer cancelCtx()

	//
	filter := bson.M{"media.fileuniqueid": uniqueID}

	//
	result := collection.FindOne(ctx, filter, options.FindOne())
	if result.Err() != nil {
		return post, result.Err()
	}

	//
	err = result.Decode(&post)
	return post, err

}

//
//// FindPostByID retrieves a post entity via its database id
//func FindPostByID(id uint) (post entities.Post) {
//
//}
//
//// DeletePostByFileID deletes a post entity via its fileID
//func DeletePostByFileID(fileID string) error {
//
//}
//
//// GetNextPost retrieves the oldest media in the queue
//func GetNextPost() (entities.Post, error) {
//
//
//}
//
//// GetQueueLength returns the number of the enqueued posts
//func GetQueueLength() (length int) {
//
//}
//
//// GetQueuePositionByDatabaseID returns the position of the selected post in the queue
//func GetQueuePositionByDatabaseID(id uint) (position int) {
//
//}
//
//// MarkPostAsPosted marks a post as posted
//func MarkPostAsPosted(post entities.Post, messageID int) error {
//
//}
//
//// MarkPostAsFailed marks a post as failed
//func MarkPostAsFailed(post entities.Post) error {
//
//
//}




// ============================================================================

func findBestMatch(referencePHash string, cursor *mongo.Cursor) (post entities.Post, err error) {

	defer func() {
		_ = cursor.Close(dsCtx)
	}()

	i := 0
	for cursor.Next(context.TODO()) {

		i++
		// Support variable. If we deserialize directly in media,
		// since IsWhitelisted is an omitempty field, it won't be
		// deserialized in case of it being missing. This way, if
		// a document with it set to true has already been retrieved,
		// it will always keep being true.
		var res entities.Post
		err = cursor.Decode(&res)
		if err == nil && fpcompare.PhotosAreSimilarEnough(referencePHash, res.Media.PHash) {
			post = res
			fmt.Println("match in ", i, "iterations. FileID", post.Media.FileUniqueID)
			return
		}

	}

	err = xerrors.New("no match found")
	return

}
