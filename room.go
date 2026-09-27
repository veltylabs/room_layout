package roomlayout

import (
	"webtyp.com/fmt"
	"webtyp.com/model"
	"webtyp.com/orm"
)

func (m *Module) SaveRoom(r Room) (Room, error) {
	if r.TenantId == "" {
		return Room{}, ErrTenantRequired
	}
	r.Code = fmt.TrimSpace(r.Code)
	r.Name = fmt.TrimSpace(r.Name)
	isNew := r.Id == ""
	if isNew {
		r.Id = m.deps.IDs.NewID()
	}

	if err := model.ValidateFields(byte(model.Create), &r); err != nil {
		return Room{}, &ValidationError{Err: err}
	}

	var codeExists bool
	err := m.db.Query(&Room{}).
		Where("tenant_id").Eq(r.TenantId).
		Where("code").Eq(r.Code).
		ReadAll(func() model.Model { return &Room{} }, func(row model.Model) {
			if existing, ok := row.(*Room); ok {
				if existing.Id != r.Id {
					codeExists = true
				}
			}
		})
	if err != nil {
		return Room{}, err
	}
	if codeExists {
		return Room{}, ErrCodeAlreadyExists
	}

	r.UpdatedAt = m.now() / 1_000_000_000
	if isNew {
		if err := m.db.Create(&r); err != nil {
			return Room{}, err
		}
	} else {
		if err := m.db.Update(&r, orm.Eq("id", r.Id), orm.Eq("tenant_id", r.TenantId)); err != nil {
			return Room{}, err
		}
	}

	m.publish(TopicRoomSaved, &r)
	return r, nil
}

func (m *Module) GetRoom(tenantID, id string) (Room, error) {
	if tenantID == "" {
		return Room{}, ErrTenantRequired
	}
	if id == "" {
		return Room{}, ErrNotFound
	}
	var res Room
	var found bool
	err := m.db.Query(&Room{}).
		Where("tenant_id").Eq(tenantID).
		Where("id").Eq(id).
		Limit(1).
		ReadAll(func() model.Model { return &Room{} }, func(row model.Model) {
			if item, ok := row.(*Room); ok {
				res = *item
				found = true
			}
		})
	if err != nil {
		return Room{}, err
	}
	if !found {
		return Room{}, ErrNotFound
	}
	return res, nil
}

func (m *Module) ListRooms(tenantID string, activeOnly bool) ([]Room, error) {
	if tenantID == "" {
		return nil, ErrTenantRequired
	}
	qb := m.db.Query(&Room{}).
		Where("tenant_id").Eq(tenantID)
	if activeOnly {
		qb = qb.Where("is_active").Eq(true)
	}
	qb = qb.OrderBy("code").Asc()

	var res []Room
	err := qb.ReadAll(func() model.Model { return &Room{} }, func(row model.Model) {
		if item, ok := row.(*Room); ok {
			res = append(res, *item)
		}
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *Module) DeactivateRoom(tenantID, id string) error {
	r, err := m.GetRoom(tenantID, id)
	if err != nil {
		return err
	}
	r.IsActive = false
	r.UpdatedAt = m.now() / 1_000_000_000
	if err := m.db.Update(&r, orm.Eq("id", id), orm.Eq("tenant_id", tenantID)); err != nil {
		return err
	}
	m.publish(TopicRoomDeactivated, &r)
	return nil
}
