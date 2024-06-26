package fs

import (
	"fmt"
	"strings"
	"time"

	"github.com/phdavis1027/go-irodsclient/irods/types"
	"github.com/phdavis1027/go-irodsclient/irods/util"
	gocache "github.com/patrickmn/go-cache"
)

// MetadataCacheTimeoutSetting defines cache timeout for path
type MetadataCacheTimeoutSetting struct {
	Path    string
	Timeout time.Duration
	Inherit bool
}

// FileSystemCache manages filesystem caches
type FileSystemCache struct {
	cacheTimeout                          time.Duration
	cleanupTimeout                        time.Duration
	cacheTimeoutPaths                     []MetadataCacheTimeoutSetting
	cacheTimeoutPathMap                   map[string]MetadataCacheTimeoutSetting
	invalidateParentEntryCacheImmediately bool
	entryCache                            *gocache.Cache
	negativeEntryCache                    *gocache.Cache
	dirCache                              *gocache.Cache
	metadataCache                         *gocache.Cache
	groupUsersCache                       *gocache.Cache
	userGroupsCache                       *gocache.Cache
	groupsCache                           *gocache.Cache
	usersCache                            *gocache.Cache
	aclCache                              *gocache.Cache
}

// NewFileSystemCache creates a new FileSystemCache
func NewFileSystemCache(cacheTimeout time.Duration, cleanup time.Duration, cacheTimeoutSettings []MetadataCacheTimeoutSetting, invalidateParentEntryCacheImmediately bool) *FileSystemCache {
	entryCache := gocache.New(cacheTimeout, cleanup)
	negativeEntryCache := gocache.New(cacheTimeout, cleanup)
	dirCache := gocache.New(cacheTimeout, cleanup)
	metadataCache := gocache.New(cacheTimeout, cleanup)
	groupUsersCache := gocache.New(cacheTimeout, cleanup)
	userGroupsCache := gocache.New(cacheTimeout, cleanup)
	groupsCache := gocache.New(cacheTimeout, cleanup)
	usersCache := gocache.New(cacheTimeout, cleanup)
	aclCache := gocache.New(cacheTimeout, cleanup)

	if cacheTimeoutSettings == nil {
		cacheTimeoutSettings = []MetadataCacheTimeoutSetting{}
	}

	// build a map for quick search
	cacheTimeoutSettingMap := map[string]MetadataCacheTimeoutSetting{}
	for _, timeoutSetting := range cacheTimeoutSettings {
		cacheTimeoutSettingMap[timeoutSetting.Path] = timeoutSetting
	}

	return &FileSystemCache{
		cacheTimeout:                          cacheTimeout,
		cleanupTimeout:                        cleanup,
		cacheTimeoutPaths:                     cacheTimeoutSettings,
		cacheTimeoutPathMap:                   cacheTimeoutSettingMap,
		invalidateParentEntryCacheImmediately: invalidateParentEntryCacheImmediately,
		entryCache:                            entryCache,
		negativeEntryCache:                    negativeEntryCache,
		dirCache:                              dirCache,
		metadataCache:                         metadataCache,
		groupUsersCache:                       groupUsersCache,
		userGroupsCache:                       userGroupsCache,
		groupsCache:                           groupsCache,
		usersCache:                            usersCache,
		aclCache:                              aclCache,
	}
}

func (cache *FileSystemCache) getCacheTTLForPath(path string) time.Duration {
	if len(cache.cacheTimeoutPathMap) == 0 {
		// no data
		return 0
	}

	// check map first
	if timeoutSetting, ok := cache.cacheTimeoutPathMap[path]; ok {
		// exact match
		return timeoutSetting.Timeout
	}

	// check inherit
	parentPaths := util.GetParentIRODSDirs(path)
	for i := len(parentPaths) - 1; i >= 0; i-- {
		parentPath := parentPaths[i]

		if timeoutSetting, ok := cache.cacheTimeoutPathMap[parentPath]; ok {
			// parent match
			if timeoutSetting.Inherit {
				// inherit
				return timeoutSetting.Timeout
			}
		}
	}

	// use default
	return 0
}

// AddEntryCache adds an entry cache
func (cache *FileSystemCache) AddEntryCache(entry *Entry) {
	ttl := cache.getCacheTTLForPath(entry.Path)
	cache.entryCache.Set(entry.Path, entry, ttl)
}

// RemoveEntryCache removes an entry cache
func (cache *FileSystemCache) RemoveEntryCache(path string) {
	cache.entryCache.Delete(path)
}

// RemoveParentDirCache removes an entry cache for the parent path of the given path
func (cache *FileSystemCache) RemoveParentDirCache(path string) {
	if cache.invalidateParentEntryCacheImmediately {
		parentPath := util.GetIRODSPathDirname(path)
		cache.entryCache.Delete(parentPath)
	}
}

// GetEntryCache retrieves an entry cache
func (cache *FileSystemCache) GetEntryCache(path string) *Entry {
	if entry, exist := cache.entryCache.Get(path); exist {
		if fsentry, ok := entry.(*Entry); ok {
			return fsentry
		}
	}
	return nil
}

// ClearEntryCache clears all entry caches
func (cache *FileSystemCache) ClearEntryCache() {
	cache.entryCache.Flush()
}

// AddNegativeEntryCache adds a negative entry cache
func (cache *FileSystemCache) AddNegativeEntryCache(path string) {
	ttl := cache.getCacheTTLForPath(path)
	cache.negativeEntryCache.Set(path, true, ttl)
}

// RemoveNegativeEntryCache removes a negative entry cache
func (cache *FileSystemCache) RemoveNegativeEntryCache(path string) {
	cache.negativeEntryCache.Delete(path)
}

// RemoveAllNegativeEntryCacheForPath removes all negative entry caches
func (cache *FileSystemCache) RemoveAllNegativeEntryCacheForPath(path string) {
	prefix := fmt.Sprintf("%s/", path)
	deleteKey := []string{}
	for k := range cache.negativeEntryCache.Items() {
		if k == path || strings.HasPrefix(k, prefix) {
			deleteKey = append(deleteKey, k)
		}
	}

	for _, k := range deleteKey {
		cache.negativeEntryCache.Delete(k)
	}
}

// HasNegativeEntryCache checks the existence of a negative entry cache
func (cache *FileSystemCache) HasNegativeEntryCache(path string) bool {
	if exist, existOk := cache.negativeEntryCache.Get(path); existOk {
		if bexist, ok := exist.(bool); ok {
			return bexist
		}
	}
	return false
}

// ClearNegativeEntryCache clears all negative entry caches
func (cache *FileSystemCache) ClearNegativeEntryCache() {
	cache.negativeEntryCache.Flush()
}

// AddDirCache adds a dir cache
func (cache *FileSystemCache) AddDirCache(path string, entries []string) {
	ttl := cache.getCacheTTLForPath(path)
	cache.dirCache.Set(path, entries, ttl)
}

// RemoveDirCache removes a dir cache
func (cache *FileSystemCache) RemoveDirCache(path string) {
	cache.dirCache.Delete(path)
}

// GetDirCache retrives a dir cache
func (cache *FileSystemCache) GetDirCache(path string) []string {
	data, exist := cache.dirCache.Get(path)
	if exist {
		if entries, ok := data.([]string); ok {
			return entries
		}
	}
	return nil
}

// ClearDirCache clears all dir caches
func (cache *FileSystemCache) ClearDirCache() {
	cache.dirCache.Flush()
}

// AddMetadataCache adds a metadata cache
func (cache *FileSystemCache) AddMetadataCache(path string, metas []*types.IRODSMeta) {
	ttl := cache.getCacheTTLForPath(path)
	cache.metadataCache.Set(path, metas, ttl)
}

// RemoveMetadataCache removes a metadata cache
func (cache *FileSystemCache) RemoveMetadataCache(path string) {
	cache.metadataCache.Delete(path)
}

// GetMetadataCache retrieves a metadata cache
func (cache *FileSystemCache) GetMetadataCache(path string) []*types.IRODSMeta {
	data, exist := cache.metadataCache.Get(path)
	if exist {
		if metas, ok := data.([]*types.IRODSMeta); ok {
			return metas
		}
	}
	return nil
}

// ClearMetadataCache clears all metadata caches
func (cache *FileSystemCache) ClearMetadataCache() {
	cache.metadataCache.Flush()
}

// AddGroupUsersCache adds a group user (users in a group) cache
func (cache *FileSystemCache) AddGroupUsersCache(group string, users []*types.IRODSUser) {
	cache.groupUsersCache.Set(group, users, 0)
}

// RemoveGroupUsersCache removes a group user (users in a group) cache
func (cache *FileSystemCache) RemoveGroupUsersCache(group string) {
	cache.groupUsersCache.Delete(group)
}

// GetGroupUsersCache retrives a group user (users in a group) cache
func (cache *FileSystemCache) GetGroupUsersCache(group string) []*types.IRODSUser {
	users, exist := cache.groupUsersCache.Get(group)
	if exist {
		if irodsUsers, ok := users.([]*types.IRODSUser); ok {
			return irodsUsers
		}
	}
	return nil
}

// AddUserGroupsCache adds a user's groups (groups that a user belongs to) cache
func (cache *FileSystemCache) AddUserGroupsCache(user string, groups []*types.IRODSUser) {
	cache.userGroupsCache.Set(user, groups, 0)
}

// RemoveUserGroupsCache removes a user's groups (groups that a user belongs to) cache
func (cache *FileSystemCache) RemoveUserGroupsCache(user string) {
	cache.userGroupsCache.Delete(user)
}

// GetUserGroupsCache retrives a user's groups (groups that a user belongs to) cache
func (cache *FileSystemCache) GetUserGroupsCache(user string) []*types.IRODSUser {
	groups, exist := cache.userGroupsCache.Get(user)
	if exist {
		if irodsGroups, ok := groups.([]*types.IRODSUser); ok {
			return irodsGroups
		}
	}
	return nil
}

// AddGroupsCache adds a groups cache (cache of a list of all groups)
func (cache *FileSystemCache) AddGroupsCache(groups []*types.IRODSUser) {
	cache.groupsCache.Set("groups", groups, 0)
}

// RemoveGroupsCache removes a groups cache (cache of a list of all groups)
func (cache *FileSystemCache) RemoveGroupsCache() {
	cache.groupsCache.Delete("groups")
}

// GetGroupsCache retrives a groups cache (cache of a list of all groups)
func (cache *FileSystemCache) GetGroupsCache() []*types.IRODSUser {
	groups, exist := cache.groupsCache.Get("groups")
	if exist {
		if irodsGroups, ok := groups.([]*types.IRODSUser); ok {
			return irodsGroups
		}
	}
	return nil
}

// AddUsersCache adds a users cache (cache of a list of all users)
func (cache *FileSystemCache) AddUsersCache(users []*types.IRODSUser) {
	cache.usersCache.Set("users", users, 0)
}

// RemoveUsersCache removes a users cache (cache of a list of all users)
func (cache *FileSystemCache) RemoveUsersCache() {
	cache.usersCache.Delete("users")
}

// GetUsersCache retrives a users cache (cache of a list of all users)
func (cache *FileSystemCache) GetUsersCache() []*types.IRODSUser {
	users, exist := cache.usersCache.Get("users")
	if exist {
		if irodsUsers, ok := users.([]*types.IRODSUser); ok {
			return irodsUsers
		}
	}
	return nil
}

// AddACLsCache adds a ACLs cache
func (cache *FileSystemCache) AddACLsCache(path string, accesses []*types.IRODSAccess) {
	ttl := cache.getCacheTTLForPath(path)
	cache.aclCache.Set(path, accesses, ttl)
}

// AddACLsCacheMulti adds multiple ACLs caches
func (cache *FileSystemCache) AddACLsCacheMulti(accesses []*types.IRODSAccess) {
	m := map[string][]*types.IRODSAccess{}

	for _, access := range accesses {
		if existingAccesses, ok := m[access.Path]; ok {
			// has it, add
			existingAccesses = append(existingAccesses, access)
			m[access.Path] = existingAccesses
		} else {
			// create it
			m[access.Path] = []*types.IRODSAccess{access}
		}
	}

	for path, access := range m {
		ttl := cache.getCacheTTLForPath(path)
		cache.aclCache.Set(path, access, ttl)
	}
}

// RemoveACLsCache removes a ACLs cache
func (cache *FileSystemCache) RemoveACLsCache(path string) {
	cache.aclCache.Delete(path)
}

// GetACLsCache retrives a ACLs cache
func (cache *FileSystemCache) GetACLsCache(path string) []*types.IRODSAccess {
	data, exist := cache.aclCache.Get(path)
	if exist {
		if entries, ok := data.([]*types.IRODSAccess); ok {
			return entries
		}
	}
	return nil
}

// ClearACLsCache clears all ACLs caches
func (cache *FileSystemCache) ClearACLsCache() {
	cache.aclCache.Flush()
}
