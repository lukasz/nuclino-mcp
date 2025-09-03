# 📋 Nuclino API Documentation & Testing Results

**Data utworzenia:** 2025-09-03  
**Autor:** Claude Code Assistant  
**Cel:** Dokumentacja oficjalnego API Nuclino na podstawie testów z prawdziwymi endpointami

---

## 🔗 **Źródła oficjalne:**

- **Główna dokumentacja API:** https://help.nuclino.com/d3a29686-api
- **Authentication:** https://help.nuclino.com/8090bb76-authentication  
- **Resources & endpoints:** https://help.nuclino.com/e30a68b1-resources-endpoints
- **Items and Collections:** https://help.nuclino.com/fa38d15f-items-and-collections
- **Manage API keys:** https://help.nuclino.com/04598850-manage-api-keys

---

## 🌐 **API Basics:**

- **Base URL:** `https://api.nuclino.com`
- **Protocol:** HTTPS only (HTTP requests fail)
- **Authentication:** `Authorization: YOUR_API_KEY` header (no "Bearer " prefix)
- **Content-Type:** `application/json`
- **Response format:** Wrapped in `{"status": "success", "data": {...}}`

---

## ✅ **WORKING ENDPOINTS - Przetestowane i działające:**

### **1. Teams**
```bash
GET /v0/teams
# Zwraca listę teamów użytkownika
```

**Response structure:**
```json
{
  "status": "success", 
  "data": {
    "object": "list",
    "results": [
      {
        "object": "team",
        "id": "team-id",
        "url": "https://app.nuclino.com/TeamName", 
        "name": "Team Name",
        "createdAt": "2025-09-01T20:10:31.775Z",
        "createdUserId": "user-id"
      }
    ]
  }
}
```

### **2. Workspaces**
```bash
GET /v0/workspaces
# Zwraca listę wszystkich workspace'ów
```

**Response structure:**
```json
{
  "status": "success",
  "data": {
    "object": "list", 
    "results": [
      {
        "object": "workspace",
        "id": "workspace-id",
        "teamId": "team-id",
        "name": "Workspace Name",
        "createdAt": "2025-09-02T18:08:27.613Z",
        "createdUserId": "user-id",
        "fields": [],
        "childIds": ["item-id-1", "item-id-2"]  // ← Items in workspace
      }
    ]
  }
}
```

### **3. Items - List**
```bash
GET /v0/items?workspaceId={workspace-id}
# Listuje wszystkie items w workspace
GET /v0/items?teamId={team-id}
# Listuje items w team (alternatywne)
```

**Response structure:**
```json
{
  "status": "success",
  "data": {
    "object": "list",
    "results": [
      {
        "object": "item",
        "id": "item-id", 
        "workspaceId": "workspace-id",
        "url": "https://app.nuclino.com/t/b/item-id",
        "title": "Document Title",
        "createdAt": "2025-09-03T11:05:40.617Z",
        "createdUserId": "user-id",
        "lastUpdatedAt": "2025-09-03T11:05:40.617Z", 
        "lastUpdatedUserId": "user-id",
        "contentMeta": {"itemIds": [], "fileIds": []},
        "fields": {}
      }
    ]
  }
}
```

### **4. Items - Get Single**
```bash
GET /v0/items/{item-id}
# Pobiera pełny item z contentem
```

**Response zawiera dodatkowo:**
- `"content": "Full markdown content here"`

### **5. Items - Search** ✅
```bash
GET /v0/items?workspaceId={workspace-id}&search={query}
# Search działa! Był błąd w naszej implementacji
```

**Response:**
- Taki sam jak list items
- Dodatkowo `"highlight"` field z highlighted fragment

### **6. Items - Update** ✅ 
```bash
PUT /v0/items/{item-id}
Content-Type: application/json
{
  "title": "New Title",
  "content": "New content"
}
```

**⚠️ WAŻNE:** Używa **PUT**, nie PATCH!

**Response:** Pełny updated item object

### **7. Items - Delete**
```bash
DELETE /v0/items/{item-id}
# Przenosi item do trash
```

---

## ❌ **NOT WORKING / PROBLEMATIC ENDPOINTS:**

### **1. Create Item - Wymaga dalszych badań**
```bash
POST /v0/items
# Zwraca 404 - może wymaga innych parametrów
```

**Próbowane warianty:**
- `{"workspaceId": "...", "title": "...", "content": "..."}` → 404
- `{"parentId": "...", "title": "...", "content": "..."}` → "Item not found"

**Możliwe przyczyny:**
- Endpoint może nie istnieć publicznie
- Może wymagać innych uprawnień
- Może potrzebować inne parametry (object: "item", index, etc.)

### **2. Generic endpoints**
- `GET /v0/items` (bez parametrów) → 400 Bad Request
- Wymaga zawsze `workspaceId` lub `teamId`

---

## 🔧 **CO NAPRAWIĆ W MCP SERVERZE:**

### **Priorytety:**

#### **1. HIGH PRIORITY - Update Item**
- Zmienić z `PATCH` na `PUT` w `client.go`
- Endpoint jest poprawny: `/v0/items/{id}`

#### **2. MEDIUM PRIORITY - Create Item**
- Zbadać dokładnie wymagania API
- Może potrzebuje dodatkowych parametrów
- Rozważyć contact z Nuclino support

#### **3. LOW PRIORITY**
- Get workspace details - sprawdzić dlaczego zwraca puste pola
- Get team details - podobnie

### **Co już działa poprawnie:**
- ✅ Authentication (bez Bearer prefix)
- ✅ List workspaces  
- ✅ List items (endpoint `/v0/items?workspaceId=...`)
- ✅ Get item details
- ✅ Delete item
- ✅ Search items (endpoint poprawny!)

---

## 📊 **TESTING SUMMARY:**

### **Test Environment:**
- API Key: Działający production key
- Test workspace: "Astras" (id: 7f2868da-dd5f-441e-9d0b-57a4381a88e0)
- Test date: 2025-09-03

### **Results:**
- **7/8 głównych funkcji działa** ✅
- **Search naprawiony** - błąd był po naszej stronie  
- **Update naprawiony** - wymagał PUT zamiast PATCH
- **Create item** - wymaga dalszych badań

### **Success Rate:** ~87.5% głównych funkcji

---

## 🚀 **NEXT STEPS:**

1. **Zaimplementować poprawki** w MCP server:
   - Update method: PATCH → PUT
   - Potwierdzić że search endpoint jest poprawny

2. **Zbadać Create Item:**
   - Sprawdzić official docs bardziej szczegółowo
   - Może contact Nuclino support

3. **Testing:**
   - Przetestować wszystko w Claude Desktop
   - Potwierdzić że wszystkie naprawki działają

---

## 📝 **CHANGELOG:**

- **2025-09-03 11:15** - Pierwsze testy API z prawdziwymi endpointami
- **2025-09-03 11:16** - Odkryto że search i update działają, błąd był w implementacji
- **2025-09-03 11:17** - Dokumentacja utworzona na podstawie real testing results

---

**🔍 Verification:** Ten dokument jest oparty na real API testing z production API key. Wszystkie endpointy były testowane curl'em przeciwko `https://api.nuclino.com`.

**⚠️ Note:** API może się zmieniać - sprawdzaj official docs przed większymi zmianami!