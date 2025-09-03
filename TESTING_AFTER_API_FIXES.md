# 🧪 Instrukcje testowania po naprawach API

**Data:** 2025-09-03  
**Status:** Gotowe do testowania w Claude Desktop

---

## 🔧 **CO ZOSTAŁO NAPRAWIONE:**

### ✅ **1. UpdateItem - zmiana PATCH → PUT**
**Było:** `PATCH /v0/items/{id}` (błąd 404)  
**Jest:** `PUT /v0/items/{id}` (zgodne z oficjalną dokumentacją)

### ✅ **2. SearchItems - zmiana POST → GET** 
**Było:** `POST /v0/items/search` (błąd 404)  
**Jest:** `GET /v0/items?workspaceId=...&search=query` (zgodne z API)

### ✅ **3. Dokumentacja API**
- Pełne testy z prawdziwymi endpointami
- Zapisana dokumentacja w `NUCLINO_API_DOCUMENTATION.md`

---

## 🚀 **RESTART WYMAGANY:**

**KROK 1: Restart Claude Desktop**
1. Zamknij Claude Desktop całkowicie
2. Poczekaj 10 sekund
3. Uruchom ponownie

---

## 🧪 **TESTY DO WYKONANIA:**

### **Test 1: Update Item (powinno teraz działać!)**
```javascript
// Znajdź item do aktualizacji
nuclino_list_items({workspace_id: "your_workspace_id"})

// Zaktualizuj go (poprzednio zwracało 404)
nuclino_update_item({
  item_id: "item_id_from_list",
  title: "Updated Title After Fix",
  content: "This update should work now with PUT method!"
})
```

### **Test 2: Search Items (powinno teraz działać!)**
```javascript
// Search który poprzednio zwracał 404
nuclino_search_items({
  workspace_id: "your_workspace_id",
  query: "test"
})

// Search z większą ilością parametrów
nuclino_search_items({
  workspace_id: "your_workspace_id", 
  query: "raport",
  limit: 5
})
```

### **Test 3: Wszystkie inne funkcje (potwierdzenie że działają)**
```javascript
// Te powinny działać tak jak wcześniej
nuclino_list_workspaces({})
nuclino_list_items({workspace_id: "your_workspace_id"})
nuclino_get_item({item_id: "some_item_id"})
nuclino_delete_item({item_id: "item_to_delete"})
```

---

## 🎯 **OCZEKIWANE REZULTATY:**

### **✅ Powinny teraz działać:**
1. **`nuclino_update_item`** - bez błędu 404
2. **`nuclino_search_items`** - bez błędu 404  
3. **Wszystkie poprzednio działające funkcje** - bez zmian

### **❌ Wciąż mogą nie działać:**
1. **`nuclino_create_item`** - wymaga dalszych badań API
2. **`nuclino_get_workspace`** - może zwracać puste pola
3. **`nuclino_create_workspace`** - może nie być dostępne publicznie

---

## 📋 **EXPECTED SUCCESS RATE:**

**Przed poprawkami:** ~62% funkcji działało  
**Po poprawkach:** **~87% funkcji powinno działać** 🎯

Główne funcje które powinny być w pełni użytkowe:
- ✅ List workspaces
- ✅ List items  
- ✅ Get item details
- ✅ **Update item** (NAPRAWIONY!)
- ✅ **Search items** (NAPRAWIONY!)
- ✅ Delete item

---

## 🐛 **Jeśli wciąż są problemy:**

### **Problem z update/search:**
1. Sprawdź logi Claude Desktop
2. Zrestartuj serwer MCP: 
   ```bash
   pkill -f nuclino-mcp-server
   ```
3. Sprawdź czy używasz najnowszej wersji binarki

### **Debug logging:**
Włącz w config jeśli potrzebujesz więcej informacji:
```json
{
  "env": {
    "LOG_LEVEL": "debug"
  }
}
```

---

## 📊 **RAPORTUJ WYNIKI:**

**Format:**
```
✅ nuclino_update_item - DZIAŁA! 
✅ nuclino_search_items - DZIAŁA!
❌ nuclino_create_item - wciąż nie działa
```

**Szczególnie ważne:**
- Czy `update_item` teraz działa (PUT vs PATCH fix)
- Czy `search_items` teraz działa (GET vs POST fix)  
- Jakiej treści błędy jeśli wciąż występują

---

**🎉 Oczekuje że teraz większość funkcji będzie działać poprawnie!**

Przetestuj i daj znać wyniki! 🚀