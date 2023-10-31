package pbf

import (
	"context"
	"fmt"
	"io/fs"
	"strconv"
	"strings"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
)

func ExtractObject(ctx context.Context, f fs.File, t string, refs []string, tagValues []string) error {
	scanner := osmpbf.New(ctx, f, 1)
	defer scanner.Close()

	lookupIDs := make(map[int64]struct{})
	for _, id := range refs {
		objectID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid object ref id: %w", err)
		}

		lookupIDs[objectID] = struct{}{}
	}

	tv := make([][]string, 0, len(tagValues))
	for _, tagValue := range tagValues {
		ss := strings.Split(tagValue, "=")
		if len(ss) != 2 {
			return fmt.Errorf("invalid tag value: %s", tagValue)
		}
		tv = append(tv, ss)
	}

	for scanner.Scan() {
		o := scanner.Object()
		if _, ok := lookupIDs[o.ObjectID().Ref()]; !ok && len(lookupIDs) > 0 {
			continue
		}

		switch o.ObjectID().Type() {
		case osm.TypeNode:
			node := o.(*osm.Node)
			if !isMatchTagValues(node.Tags, tv) {
				continue
			}
			fmt.Println(o.ObjectID().Ref(), node.TagMap())
		case osm.TypeWay:
			way := o.(*osm.Way)
			if !isMatchTagValues(way.Tags, tv) {
				continue
			}
			fmt.Println(o.ObjectID().Ref(), way.TagMap())
		case osm.TypeRelation:
			relation := o.(*osm.Relation)
			if !isMatchTagValues(relation.Tags, tv) {
				continue
			}
			fmt.Println(o.ObjectID().Ref(), relation.TagMap())
		}
	}
	return nil
}

func isMatchTagValues(tags osm.Tags, tv [][]string) bool {
	if len(tv) == 0 {
		return true
	}

	for _, t := range tv {
		if tags.Find(t[0]) != t[1] {
			return false
		}
	}
	return true
}
