# 🧪 Instrukcja testowania po poprawkach

## Poprawione problemy:

### ✅ **1. CreateItemTool - naprawione parametry**
- **BYŁO:** Wymagał `collection_id` 
- **JEST:** Wymaga `workspace_id` i opcjonalnie `parent_id`

**Test:**
```javascript
nuclino_create_item({
  workspace_id: "your_workspace_id",  // NOWY parametr
  title: "Test Document", 
  content: "This is test content"
})
```

### ✅ **2. Wyłączone problematyczne tool'e**
Czasowo wyłączyłem tool'e które prawdopodobnie nie działają z Nuclino API:
- `nuclino_list_collections` 
- `nuclino_create_collection`
- `nuclino_move_item`
- `nuclino_list_collection_items`
- Wszystkie collection-related tools

**Test:** Te tool'e nie powinny już być widoczne w Claude Desktop.

### ✅ **3. Naprawiony endpoint dla ListItems**  
- **BYŁO:** `GET /v0/workspaces/{id}/items`
- **JEST:** `GET /v0/items?workspaceId={id}`

**Test:**
```javascript
nuclino_list_items({
  workspace_id: "your_workspace_id"
})
```

---

## 🔧 Restartuj Claude Desktop!

**WAŻNE:** Żeby zobaczyć zmiany:
1. Zamknij Claude Desktop całkowicie
2. Poczekaj 10 sekund  
3. Uruchom ponownie Claude Desktop

---

## 🧪 Testy do wykonania:

### **Test 1: Workspace operations (powinny działać)**
```javascript
// Test podstawowy
nuclino_list_workspaces({})

// Test szczegółów workspace  
nuclino_get_workspace({workspace_id: "your_workspace_id"})
```

### **Test 2: Item operations** 
```javascript
// Test listowania (nowy endpoint)
nuclino_list_items({workspace_id: "your_workspace_id"})

// Test wyszukiwania
nuclino_search_items({query: "test", workspace_id: "your_workspace_id"})

// Test tworzenia (nowe parametry)
nuclino_create_item({
  workspace_id: "your_workspace_id",
  title: "Test Document Claude",
  content: "Test content created from Claude Desktop"
})
```

### **Test 3: Team operations**
```javascript
nuclino_list_teams({})
nuclino_get_team({team_id: "your_team_id"})
```

---

## 📋 Sprawdź te rzeczy:

### ✅ **Powinno działać:**
1. **`nuclino_list_workspaces`** - już działało
2. **`nuclino_create_item`** - z nowym `workspace_id` parametrem
3. **`nuclino_list_items`** - z nowym endpointem 
4. **`nuclino_search_items`** - powinno działać
5. **`nuclino_get_item`** - jeśli masz item ID

### ❓ **Do sprawdzenia:**
1. **`nuclino_get_workspace`** - czy zwraca pełne dane?
2. **`nuclino_get_team`** - czy zwraca pełne dane?
3. **`nuclino_list_teams`** - czy działa?

### ❌ **Powinno być wyłączone:**
- Wszystkie `collection_*` tools
- `nuclino_move_item`  
- `nuclino_list_collection_items`

---

## 🐛 Jeśli wciąż są problemy:

### **Puste dane w get_workspace/get_team:**
Jeśli `get_workspace` zwraca pustą strukturę, to prawdopodobnie:
- Response ma inną strukturę niż oczekuję
- Pola mają inne nazwy w JSON

### **404 errors na innych endpointach:**
- `list_items` - sprawdź czy nowy endpoint `/v0/items?workspaceId=...` działa
- Jeśli nie, to może Nuclino API ma zupełnie inną strukturę

### **Inne błędy:**
Włącz debug logging w Claude Desktop config:
```json
{
  "mcpServers": {
    "nuclino": {
      "command": "/path/to/nuclino-mcp-server",
      "env": {
        "NUCLINO_API_KEY": "your_key",
        "LOG_LEVEL": "debug"
      }
    }
  }
}
```

---

## 📬 Raportuj wyniki:

**Podaj:**
1. Które tool'e teraz działają ✅
2. Które wciąż zwracają błędy ❌ 
3. Jakiej treści błędy
4. Czy `create_item` z `workspace_id` działa
5. Czy `list_items` z nowym endpointem działa

**Przykład raportu:**
```
✅ nuclino_create_item - działa z workspace_id!
❌ nuclino_list_items - wciąż 404
❌ nuclino_get_workspace - zwraca puste pola
```