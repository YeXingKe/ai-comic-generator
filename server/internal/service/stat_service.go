package service

import (
	"math"
	"sort"
	"time"

	"github.com/ai-comic-generator/server/internal/model"
	"github.com/ai-comic-generator/server/internal/store"
	"golang.org/x/sync/errgroup"
)

type StatService struct {
	statStore *store.StatStore
}

func NewStatService(statStore *store.StatStore) *StatService {
	return &StatService{statStore: statStore}
}

func (s *StatService) GetDashboard(req *model.StatQueryRequest) (*model.StatDashboard, error) {
	now := time.Now()
	days := req.Range.ToDays()
	startTime := now.AddDate(0, 0, -days)
	prevStart := startTime.AddDate(0, 0, -days)

	var (
		totalComics     int64
		completedComics int64
		avgDuration     float64
		totalUsers      int64
		weeklyNew       int64
		trend           []model.StatTrendPoint
		statusDist      []model.StatBucket
		phaseFunnel     []model.StatBucket
		styleDist       []model.StatBucket
		roleDist        []model.StatBucket
		publishDist     []model.StatBucket
		prevTotal       int64
		prevCompleted   int64
		prevWeekly      int64
	)

	g := new(errgroup.Group)

	g.Go(func() error { var e error; totalComics, e = s.statStore.CountComics(startTime); return e })
	g.Go(func() error { var e error; completedComics, e = s.statStore.CountCompletedComics(startTime); return e })
	g.Go(func() error { var e error; avgDuration, e = s.statStore.AvgDurationMin(startTime); return e })
	g.Go(func() error { var e error; totalUsers, e = s.statStore.CountUsers(); return e })
	g.Go(func() error { var e error; weeklyNew, e = s.statStore.CountRecentComics(now.AddDate(0, 0, -7)); return e })
	g.Go(func() error { var e error; trend, e = s.statStore.GetTrend(startTime); return e })
	g.Go(func() error { var e error; statusDist, e = s.statStore.GetStatusDistribution(startTime); return e })
	g.Go(func() error { var e error; phaseFunnel, e = s.statStore.GetPhaseFunnel(startTime); return e })
	g.Go(func() error { var e error; styleDist, e = s.statStore.GetStyleDistribution(startTime); return e })
	g.Go(func() error { var e error; roleDist, e = s.statStore.GetRoleDistribution(); return e })
	g.Go(func() error { var e error; publishDist, e = s.statStore.GetPublishDistribution(startTime); return e })
	g.Go(func() error { var e error; prevTotal, e = s.statStore.CountComics(prevStart); return e })
	g.Go(func() error { var e error; prevCompleted, e = s.statStore.CountCompletedComics(prevStart); return e })
	g.Go(func() error { var e error; prevWeekly, e = s.statStore.CountRecentComics(now.AddDate(0, 0, -14)); return e })

	if err := g.Wait(); err != nil {
		return nil, err
	}

	var completionRate, prevCompletionRate float64
	if totalComics > 0 {
		completionRate = float64(completedComics) / float64(totalComics)
	}
	if prevTotal > 0 {
		prevCompletionRate = float64(prevCompleted) / float64(prevTotal)
	}

	return &model.StatDashboard{
		Overview: model.StatOverview{
			TotalComics:     int(totalComics),
			CompletedComics: int(completedComics),
			CompletionRate:  roundFloat(completionRate, 4),
			TotalUsers:      int(totalUsers),
			WeeklyNewComics: int(weeklyNew),
			AvgDurationMin:  roundFloat(avgDuration, 1),
			Deltas: model.StatDeltas{
				TotalComics:     calcDelta(totalComics, prevTotal),
				CompletionRate:  calcDeltaF(completionRate, prevCompletionRate),
				TotalUsers:      0, // 全量用户无上一期对比，固定0
				WeeklyNewComics: calcDelta(weeklyNew, prevWeekly),
			},
		},
		Trend:               fillTrendGaps(trend, startTime, now),
		StatusDistribution:  applyLabels(statusDist, statusLabels),
		PhaseFunnel:         sortPhase(applyLabels(phaseFunnel, phaseLabels)),
		StyleDistribution:   applyLabels(styleDist, styleLabels),
		RoleDistribution:    applyLabels(roleDist, roleLabels),
		PublishDistribution: applyLabels(publishDist, publishLabels),
	}, nil
}

func fillTrendGaps(points []model.StatTrendPoint, start, end time.Time) []model.StatTrendPoint {
	byDate := make(map[string]model.StatTrendPoint, len(points))
	for _, p := range points {
		byDate[p.Date] = p
	}
	var result []model.StatTrendPoint
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		if p, ok := byDate[key]; ok {
			result = append(result, p)
		} else {
			result = append(result, model.StatTrendPoint{Date: key})
		}
	}
	return result
}

func applyLabels(buckets []model.StatBucket, labels map[string]string) []model.StatBucket {
	result := make([]model.StatBucket, len(buckets))
	for i, b := range buckets {
		label := b.Key
		if l, ok := labels[b.Key]; ok {
			label = l
		}
		result[i] = model.StatBucket{Key: b.Key, Label: label, Value: b.Value}
	}
	return result
}

var phaseOrder = []string{
	model.ComicPhasePending, model.ComicPhaseTitleGeneration, model.ComicPhaseTitleSelecting,
	model.ComicPhaseStoryIdeation, model.ComicPhaseCharacterDesign, model.ComicPhaseStoryboardScript,
	model.ComicPhaseImageGeneration, model.ComicPhaseLayoutCompose, model.ComicPhaseWechatPublish,
}

var phaseIndex = func() map[string]int {
	m := make(map[string]int, len(phaseOrder))
	for i, p := range phaseOrder {
		m[p] = i
	}
	return m
}()

func sortPhase(buckets []model.StatBucket) []model.StatBucket {
	sort.Slice(buckets, func(i, j int) bool {
		return phaseIndex[buckets[i].Key] < phaseIndex[buckets[j].Key]
	})
	return buckets
}

func calcDelta(curr, prev int64) float64 {
	if prev == 0 {
		return 0
	}
	return roundFloat((float64(curr)-float64(prev))/float64(prev)*100, 1)
}

func calcDeltaF(curr, prev float64) float64 {
	if prev == 0 {
		return 0
	}
	return roundFloat((curr-prev)/prev*100, 1)
}

func roundFloat(val float64, precision int) float64 {
	p := math.Pow(10, float64(precision))
	return math.Round(val*p) / p
}

var statusLabels = map[string]string{
	model.ComicStatusPending: "待开始", model.ComicStatusProcessing: "生成中",
	model.ComicStatusAwaitingConfirm: "待确认", model.ComicStatusTitleConfirmed: "已确认",
	model.ComicStatusCompleted: "已完成", model.ComicStatusFailed: "已失败",
}

var phaseLabels = map[string]string{
	model.ComicPhasePending: "初始", model.ComicPhaseTitleGeneration: "标题生成",
	model.ComicPhaseTitleSelecting: "标题选择", model.ComicPhaseStoryIdeation: "故事构思",
	model.ComicPhaseCharacterDesign: "角色设定", model.ComicPhaseStoryboardScript: "分镜脚本",
	model.ComicPhaseImageGeneration: "图片生成", model.ComicPhaseLayoutCompose: "排版合成",
	model.ComicPhaseWechatPublish: "公众号发布",
}

var styleLabels = map[string]string{
	"cartoon": "卡通风格", "realistic": "写实风格", "chibi": "Q版风格", "animal": "动物风格",
}

var roleLabels = map[string]string{
	"user": "普通用户", "admin": "管理员", "vip": "VIP用户",
}

var publishLabels = map[string]string{
	"published": "已发布", "draft": "草稿", "failed": "发布失败",
}
