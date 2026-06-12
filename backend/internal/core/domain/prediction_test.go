package domain

import "testing"

func TestPrediction_PointsEarned(t *testing.T) {
	type args struct {
		actualHome int
		actualAway int
	}
	cases := []struct {
		name string
		pred Prediction
		args args
		want int
	}{
		{
			name: "regresion: pronosticar visitante NO es pronosticar empate",
			pred: Prediction{HomeScore: 1, AwayScore: 2},
			args: args{actualHome: 1, actualAway: 1},
			want: 0,
		},
		{
			name: "regresion: pronosticar local NO es pronosticar empate",
			pred: Prediction{HomeScore: 2, AwayScore: 1},
			args: args{actualHome: 1, actualAway: 1},
			want: 0,
		},
		{
			name: "acierto exacto vale 3",
			pred: Prediction{HomeScore: 1, AwayScore: 1},
			args: args{actualHome: 1, actualAway: 1},
			want: 3,
		},
		{
			name: "acierto exacto con marcador mas alto vale 3",
			pred: Prediction{HomeScore: 3, AwayScore: 2},
			args: args{actualHome: 3, actualAway: 2},
			want: 3,
		},
		{
			name: "pronostico empate y resultado empate vale 2",
			pred: Prediction{HomeScore: 0, AwayScore: 0},
			args: args{actualHome: 2, actualAway: 2},
			want: 2,
		},
		{
			name: "pronostico victoria local y resultado victoria local vale 2",
			pred: Prediction{HomeScore: 2, AwayScore: 1},
			args: args{actualHome: 3, actualAway: 0},
			want: 2,
		},
		{
			name: "pronostico victoria visitante y resultado victoria visitante vale 2",
			pred: Prediction{HomeScore: 0, AwayScore: 1},
			args: args{actualHome: 0, actualAway: 3},
			want: 2,
		},
		{
			name: "pronostico victoria local y resultado victoria visitante vale 0",
			pred: Prediction{HomeScore: 3, AwayScore: 0},
			args: args{actualHome: 0, actualAway: 2},
			want: 0,
		},
		{
			name: "pronostico empate y resultado victoria local vale 0",
			pred: Prediction{HomeScore: 1, AwayScore: 1},
			args: args{actualHome: 2, actualAway: 0},
			want: 0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := c.pred
			if got := p.PointsEarned(c.args.actualHome, c.args.actualAway); got != c.want {
				t.Errorf("PointsEarned(%d-%d) con real %d-%d = %d, want %d",
					p.HomeScore, p.AwayScore,
					c.args.actualHome, c.args.actualAway,
					got, c.want)
			}
		})
	}
}
