package teams

import (
	cpov1 "github.com/cybertec-postgresql/cybertec-pg-operator/pkg/apis/cpo.opensource.cybertec.at/v1"
	"github.com/cybertec-postgresql/cybertec-pg-operator/pkg/util"
)

// PostgresTeamMap is the operator's internal representation of all PostgresTeam CRDs
type PostgresTeamMap map[string]postgresTeamMembership

type postgresTeamMembership struct {
	AdditionalSuperuserTeams []string
	AdditionalTeams          []string
	AdditionalMembers        []string
}

type teamHashSet map[string]map[string]struct{}

func (ths *teamHashSet) has(team string) bool {
	_, ok := (*ths)[team]
	return ok
}

func (ths *teamHashSet) add(newTeam string, newSet []string) {
	set := make(map[string]struct{})
	if ths.has(newTeam) {
		set = (*ths)[newTeam]
	}
	for _, t := range newSet {
		set[t] = struct{}{}
	}
	(*ths)[newTeam] = set
}

func (ths *teamHashSet) toMap() map[string][]string {
	newTeamMap := make(map[string][]string)
	for team, items := range *ths {
		list := []string{}
		for item := range items {
			list = append(list, item)
		}
		newTeamMap[team] = list
	}
	return newTeamMap
}

func (ths *teamHashSet) mergeCrdMap(crdTeamMap map[string][]string) {
	for t, at := range crdTeamMap {
		ths.add(t, at)
	}
}

func fetchTeams(teamset *map[string]struct{}, set teamHashSet) {
	for key := range set {
		(*teamset)[key] = struct{}{}
	}
}

func (ptm *PostgresTeamMap) fetchAdditionalTeams(team string, superuserTeams bool, transitive bool, exclude []string) []string {

	var teams []string

	if superuserTeams {
		teams = (*ptm)[team].AdditionalSuperuserTeams
	} else {
		teams = (*ptm)[team].AdditionalTeams
	}
	if transitive {
		for _, additionalTeam := range teams {
			if !(util.SliceContains(exclude, additionalTeam)) {
				// remember to not check team and additionalTeam again
				exclude = append(exclude, additionalTeam)
				transitiveTeams := (*ptm).fetchAdditionalTeams(additionalTeam, superuserTeams, transitive, exclude)
				for _, transitiveTeam := range transitiveTeams {
					if !(util.SliceContains(exclude, transitiveTeam)) {
						// remember to not check transitive team again in case
						// it is one of the next additional teams of the outer loop
						exclude = append(exclude, transitiveTeam)
						if !(util.SliceContains(teams, transitiveTeam)) {
							// found a new transitive additional team
							teams = append(teams, transitiveTeam)
						}
					}
				}
			}
		}
	}

	return teams
}

// GetAdditionalTeams function to retrieve list of additional teams
func (ptm *PostgresTeamMap) GetAdditionalTeams(team string, transitive bool) []string {
	return ptm.fetchAdditionalTeams(team, false, transitive, []string{team})
}

// GetAdditionalSuperuserTeams function to retrieve list of additional superuser teams
func (ptm *PostgresTeamMap) GetAdditionalSuperuserTeams(team string, transitive bool) []string {
	return ptm.fetchAdditionalTeams(team, true, transitive, []string{team})
}

// Load function to import data from PostgresTeam CRD
func (ptm *PostgresTeamMap) Load(pgTeams *cpov1.PostgresTeamList) {
	// reset the team map
	*ptm = make(PostgresTeamMap, 0)

	superuserTeamSet := teamHashSet{}
	teamSet := teamHashSet{}
	teamMemberSet := teamHashSet{}
	teamIDs := make(map[string]struct{})

	for _, pgTeam := range pgTeams.Items {
		superuserTeamSet.mergeCrdMap(pgTeam.Spec.AdditionalSuperuserTeams)
		teamSet.mergeCrdMap(pgTeam.Spec.AdditionalTeams)
		teamMemberSet.mergeCrdMap(pgTeam.Spec.AdditionalMembers)
	}
	fetchTeams(&teamIDs, superuserTeamSet)
	fetchTeams(&teamIDs, teamSet)
	fetchTeams(&teamIDs, teamMemberSet)

	for teamID := range teamIDs {
		(*ptm)[teamID] = postgresTeamMembership{
			AdditionalSuperuserTeams: util.CoalesceStrArr(superuserTeamSet.toMap()[teamID], []string{}),
			AdditionalTeams:          util.CoalesceStrArr(teamSet.toMap()[teamID], []string{}),
			AdditionalMembers:        util.CoalesceStrArr(teamMemberSet.toMap()[teamID], []string{}),
		}
	}
}
