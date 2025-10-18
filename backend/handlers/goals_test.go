package handlers

import "testing"

func TestBuildPlan(t *testing.T) {
	g := Goal{Title:"Test", TargetAmount: 120000, DeadlineMonths: 12, SavingsRate: 0.05}
	plan := buildPlan(g, "ru")
	if plan.Months != 12 || plan.MonthlyAmount != 10000 {
		t.Fatalf("unexpected plan: %+v", plan)
	}
	if len(plan.Steps) != 12 {
		t.Fatalf("steps len: %d", len(plan.Steps))
	}
	if len(plan.Tips) == 0 {
		t.Fatalf("expected tips")
	}
}
