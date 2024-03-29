package teams

import (
	"testing"

	cpov1 "github.com/cybertec-postgresql/cybertec-pg-operator/pkg/apis/cpo.opensource.cybertec.at/v1"
	"github.com/cybertec-postgresql/cybertec-pg-operator/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	pgTeamList = cpov1.PostgresTeamList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: "v1",
		},
		Items: []cpov1.PostgresTeam{
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PostgresTeam",
					APIVersion: "cpo.opensource.cybertec.at/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "teamAB",
				},
				Spec: cpov1.PostgresTeamSpec{
					AdditionalSuperuserTeams: map[string][]string{"teamA": []string{"teamB", "team24x7"}, "teamB": []string{"teamA", "teamC", "team24x7"}},
					AdditionalTeams:          map[string][]string{"teamA": []string{"teamC"}, "teamB": []string{}},
					AdditionalMembers:        map[string][]string{"team24x7": []string{"optimusprime"}, "teamB": []string{"drno"}},
				},
			}, {
				TypeMeta: metav1.TypeMeta{
					Kind:       "PostgresTeam",
					APIVersion: "cpo.opensource.cybertec.at/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "teamC",
				},
				Spec: cpov1.PostgresTeamSpec{
					AdditionalSuperuserTeams: map[string][]string{"teamC": []string{"team24x7"}},
					AdditionalTeams:          map[string][]string{"teamA": []string{"teamC"}, "teamC": []string{"teamA", "teamB", "acid"}},
					AdditionalMembers:        map[string][]string{"acid": []string{"batman"}},
				},
			},
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PostgresTeam",
					APIVersion: "cpo.opensource.cybertec.at/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "teamD",
				},
				Spec: cpov1.PostgresTeamSpec{
					AdditionalSuperuserTeams: map[string][]string{},
					AdditionalTeams:          map[string][]string{"teamA": []string{"teamD"}, "teamC": []string{"teamD"}, "teamD": []string{"teamA", "teamB", "teamC"}},
					AdditionalMembers:        map[string][]string{"acid": []string{"batman"}},
				},
			},
		},
	}
	pgTeamMap = PostgresTeamMap{
		"teamA": {
			AdditionalSuperuserTeams: []string{"teamB", "team24x7"},
			AdditionalTeams:          []string{"teamC", "teamD"},
			AdditionalMembers:        []string{},
		},
		"teamB": {
			AdditionalSuperuserTeams: []string{"teamA", "teamC", "team24x7"},
			AdditionalTeams:          []string{},
			AdditionalMembers:        []string{"drno"},
		},
		"teamC": {
			AdditionalSuperuserTeams: []string{"team24x7"},
			AdditionalTeams:          []string{"teamA", "teamB", "teamD", "acid"},
			AdditionalMembers:        []string{},
		},
		"teamD": {
			AdditionalSuperuserTeams: []string{},
			AdditionalTeams:          []string{"teamA", "teamB", "teamC"},
			AdditionalMembers:        []string{},
		},
		"team24x7": {
			AdditionalSuperuserTeams: []string{},
			AdditionalTeams:          []string{},
			AdditionalMembers:        []string{"optimusprime"},
		},
		"acid": {
			AdditionalSuperuserTeams: []string{},
			AdditionalTeams:          []string{},
			AdditionalMembers:        []string{"batman"},
		},
	}
)

// TestLoadingPostgresTeamCRD PostgresTeamMap is the operator's internal representation of all PostgresTeam CRDs
func TestLoadingPostgresTeamCRD(t *testing.T) {
	tests := []struct {
		name  string
		crd   cpov1.PostgresTeamList
		ptm   PostgresTeamMap
		error string
	}{
		{
			"Check that CRD is imported correctly into the internal format",
			pgTeamList,
			pgTeamMap,
			"Mismatch between PostgresTeam CRD and internal map",
		},
	}

	for _, tt := range tests {
		postgresTeamMap := PostgresTeamMap{}
		postgresTeamMap.Load(&tt.crd)
		for team, ptmeamMembership := range postgresTeamMap {
			if !util.IsEqualIgnoreOrder(ptmeamMembership.AdditionalSuperuserTeams, tt.ptm[team].AdditionalSuperuserTeams) {
				t.Errorf("%s: %v: expected additional members %#v, got %#v", tt.name, tt.error, tt.ptm, postgresTeamMap)
			}
			if !util.IsEqualIgnoreOrder(ptmeamMembership.AdditionalTeams, tt.ptm[team].AdditionalTeams) {
				t.Errorf("%s: %v: expected additional teams %#v, got %#v", tt.name, tt.error, tt.ptm, postgresTeamMap)
			}
			if !util.IsEqualIgnoreOrder(ptmeamMembership.AdditionalMembers, tt.ptm[team].AdditionalMembers) {
				t.Errorf("%s: %v: expected additional superuser teams %#v, got %#v", tt.name, tt.error, tt.ptm, postgresTeamMap)
			}
		}
	}
}

// TestGetAdditionalTeams if returns teams with and without transitive dependencies
func TestGetAdditionalTeams(t *testing.T) {
	tests := []struct {
		name       string
		team       string
		transitive bool
		teams      []string
		error      string
	}{
		{
			"Check that additional teams are returned",
			"teamA",
			false,
			[]string{"teamC", "teamD"},
			"GetAdditionalTeams returns wrong list",
		},
		{
			"Check that additional teams are returned incl. transitive teams",
			"teamA",
			true,
			[]string{"teamC", "teamD", "teamB", "acid"},
			"GetAdditionalTeams returns wrong list",
		},
		{
			"Check that empty list is returned",
			"teamB",
			false,
			[]string{},
			"GetAdditionalTeams returns wrong list",
		},
	}

	postgresTeamMap := PostgresTeamMap{}
	postgresTeamMap.Load(&pgTeamList)

	for _, tt := range tests {
		additionalTeams := postgresTeamMap.GetAdditionalTeams(tt.team, tt.transitive)
		if !util.IsEqualIgnoreOrder(additionalTeams, tt.teams) {
			t.Errorf("%s: %v: expected additional teams %#v, got %#v", tt.name, tt.error, tt.teams, additionalTeams)
		}
	}
}

// TestGetAdditionalSuperuserTeams if returns teams with and without transitive dependencies
func TestGetAdditionalSuperuserTeams(t *testing.T) {
	tests := []struct {
		name       string
		team       string
		transitive bool
		teams      []string
		error      string
	}{
		{
			"Check that additional superuser teams are returned",
			"teamA",
			false,
			[]string{"teamB", "team24x7"},
			"GetAdditionalSuperuserTeams returns wrong list",
		},
		{
			"Check that additional superuser teams are returned incl. transitive superuser teams",
			"teamA",
			true,
			[]string{"teamB", "teamC", "team24x7"},
			"GetAdditionalSuperuserTeams returns wrong list",
		},
		{
			"Check that empty list is returned",
			"team24x7",
			false,
			[]string{},
			"GetAdditionalSuperuserTeams returns wrong list",
		},
	}

	postgresTeamMap := PostgresTeamMap{}
	postgresTeamMap.Load(&pgTeamList)

	for _, tt := range tests {
		additionalTeams := postgresTeamMap.GetAdditionalSuperuserTeams(tt.team, tt.transitive)
		if !util.IsEqualIgnoreOrder(additionalTeams, tt.teams) {
			t.Errorf("%s: %v: expected additional teams %#v, got %#v", tt.name, tt.error, tt.teams, additionalTeams)
		}
	}
}
