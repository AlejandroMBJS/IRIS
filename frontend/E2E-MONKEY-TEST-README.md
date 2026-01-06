# 🐵 IRIS Monkey/Crawler E2E Test

## Descripción

Test E2E automatizado que navega por la aplicación IRIS como un usuario humano, clickeando elementos interactivos, llenando formularios y explorando páginas, pero **excluyendo automáticamente cualquier acción peligrosa o destructiva**.

Este test está diseñado para:
- ✅ Detectar errores de navegación y UI
- ✅ Validar que no haya errores 5xx en rutas comunes
- ✅ Verificar que elementos interactivos funcionen correctamente
- ✅ Probar flujos de usuario de manera exploratoria
- ❌ **NUNCA** ejecutar acciones destructivas (delete, admin, payments, etc.)

---

## 🛡️ Garantías de Seguridad

El test incluye múltiples capas de protección:

### 1. **Validación de Entorno**
- ❌ Solo se ejecuta en `localhost` o `127.0.0.1`
- ❌ Aborta automáticamente si detecta baseURL no-local

### 2. **Bloqueo de Endpoints Peligrosos**
El test intercepta y bloquea automáticamente:
- ❌ **TODOS** los métodos `DELETE`
- ❌ Rutas de admin: `/admin/*`, `/permissions/*`, `/role-inheritance`
- ❌ Endpoints de pagos: `/payment`, `/billing`, `/reimburse`, `/mark-paid`, `/invoice`
- ❌ Acciones destructivas: `/terminate`, `/decline`, `/reject`, `/archive`, `/revoke`, `/destroy`
- ❌ Operaciones sensibles: `/drop`, `/purge`, `/reset`, `/seed`, `/migrate`
- ❌ Webhooks y secrets: `/webhook`, `/secret`, `/key`, `/oauth`

### 3. **Bloqueo de Rutas Frontend Peligrosas**
Evita navegar a:
- ❌ `/admin/*`
- ❌ `/configuration/role-inheritance`
- ❌ `/configuration/permissions`
- ❌ `/billing`, `/payment`, `/subscription`

### 4. **Bloqueo de Elementos UI Peligrosos**
No clickea botones/links que contengan:
- ❌ Texto: `delete`, `remove`, `eliminar`, `borrar`, `destroy`, `drop`, `purge`, `reset`
- ❌ Texto: `pagar`, `payment`, `subscribe`, `cancel subscription`, `refund`
- ❌ Clases CSS: `danger`, `destructive`, `bg-red-*`, `text-red-*`
- ❌ `aria-label` o `data-testid` con patrones peligrosos

### 5. **Protección de Sesión**
- ❌ Evita botones de `logout` para no terminar la sesión durante el test

### 6. **Sin Envío de Formularios**
- ✅ Llena formularios con datos dummy para probar validación
- ❌ **NO** envía formularios automáticamente (evita crear/modificar datos)

---

## 📋 Requisitos Previos

### 1. Instalar Playwright

```bash
cd frontend
npm install
npx playwright install
```

### 2. Configurar Variables de Entorno (Opcional)

Para pruebas autenticadas, crea un archivo `.env.test` en `/frontend`:

```env
# Credenciales de prueba (usuario de test, NO producción)
E2E_EMAIL=test@example.com
E2E_PASSWORD=TestPassword123!

# URL base (por defecto: http://localhost:3000)
BASE_URL=http://localhost:3000

# URL del backend (por defecto: http://localhost:8080)
BACKEND_URL=http://localhost:8080
```

**⚠️ IMPORTANTE:**
- Usa SOLO credenciales de entorno local/dev
- NUNCA uses credenciales de producción
- El test validará que `BASE_URL` sea localhost

### 3. Iniciar la Aplicación

Asegúrate de que tanto frontend como backend estén corriendo:

```bash
# Terminal 1: Backend
cd backend
go run cmd/api/main.go

# Terminal 2: Frontend
cd frontend
npm run dev

# O con Docker
docker-compose up
```

Verifica que la app esté accesible en:
- Frontend: http://localhost:3000
- Backend: http://localhost:8080

---

## 🚀 Uso

### Modo Headless (recomendado para CI/CD)
```bash
npm run test:e2e
```

### Modo Headed (ver navegador)
```bash
npm run test:e2e:headed
```

### Modo Debug (paso a paso)
```bash
npm run test:e2e:debug
```

### Modo UI (interfaz interactiva de Playwright)
```bash
npm run test:e2e:ui
```

### Ver Reporte HTML
```bash
npm run test:e2e:report
```

---

## 📊 Interpretar el Reporte

Al finalizar, el test imprime un reporte en consola:

```
================================================================================
🐵 MONKEY TEST REPORT
================================================================================

📊 STATISTICS:
  Pages visited: 12
  Clicks executed: 45
  Clicks blocked (dangerous): 8
  Requests blocked: 3
  Errors detected: 0

📄 PAGES VISITED:
  1. http://localhost:3000/
  2. http://localhost:3000/dashboard
  3. http://localhost:3000/employees
  4. http://localhost:3000/payroll
  ...

🛡️ BLOCKED ACTIONS (dangerous):
  1. [CLICK_BLOCKED] Dangerous UI element detected
     Target: BUTTON:Eliminar empleado:/employees/delete
  2. [REQUEST_BLOCKED] Dangerous DELETE request
     Target: http://localhost:8080/api/v1/employees/123
  ...

❌ ERRORS:
  (lista de errores detectados, si los hay)

================================================================================
```

### Métricas Clave

| Métrica | Descripción | Valor Esperado |
|---------|-------------|----------------|
| **Pages visited** | Número de páginas únicas visitadas | ≥ 3 |
| **Clicks executed** | Clicks realizados exitosamente | ≥ 5 |
| **Clicks blocked** | Clicks bloqueados por seguridad | Variable (esperado) |
| **Requests blocked** | Requests HTTP bloqueados | Variable (esperado) |
| **Errors detected** | Errores de página/consola/5xx | **0** (ideal) |

### Interpretación

✅ **Test Exitoso:**
- Sin errores de página/consola
- Sin respuestas 5xx
- Navegación fluida sin crashes

⚠️ **Revisar si:**
- `Errors detected > 0` → Investigar errores en el reporte
- `Clicks executed < 5` → Posible problema de acceso/autenticación
- `Pages visited < 3` → Navegación bloqueada o rutas inaccesibles

❌ **Test Fallido:**
- Errores de JavaScript no capturados
- Respuestas 5xx del servidor
- Crashes de navegador

---

## ⚙️ Configuración Avanzada

### Ajustar Límites de Exploración

Edita `/frontend/tests/monkey.spec.ts`:

```typescript
const CONFIG = {
  MAX_STEPS: 100,           // Máximo de acciones antes de parar
  MAX_PAGES: 20,            // Máximo de páginas diferentes a visitar
  MAX_CLICKS_PER_PAGE: 10,  // Máximo de clicks por página
  NAVIGATION_DELAY: 500,    // ms entre acciones
  FORM_FILL_PROBABILITY: 0.7, // 70% de probabilidad de llenar formularios
};
```

### Personalizar Denylists

Agrega patrones adicionales en el archivo de test:

```typescript
const DANGEROUS_ENDPOINT_PATTERNS = [
  /\/admin\//i,
  /\/delete/i,
  // Agrega tus patrones aquí
  /\/custom-dangerous-route/i,
];
```

### Cambiar Variables de Entorno en Runtime

```bash
BASE_URL=http://localhost:3001 npm run test:e2e
E2E_EMAIL=user@test.com E2E_PASSWORD=pass npm run test:e2e
```

---

## 🔍 Casos de Uso

### 1. Smoke Test Post-Deploy
```bash
# Después de desplegar a dev/staging local
npm run test:e2e
```

### 2. Regression Testing
```bash
# Antes de cada release, validar navegación básica
npm run test:e2e
```

### 3. Exploración de Nuevas Features
```bash
# Ver en tiempo real qué clickea el test
npm run test:e2e:headed
```

### 4. CI/CD Pipeline
```yaml
# .github/workflows/e2e.yml (ejemplo)
- name: Run E2E Monkey Test
  run: |
    npm run dev &
    sleep 5
    npm run test:e2e
```

---

## 🐛 Troubleshooting

### Error: "Tests can only run on localhost"
**Causa:** `BASE_URL` no es localhost
**Solución:** Verifica que `BASE_URL=http://localhost:3000`

### Error: "No clickable elements found"
**Causa:** Página requiere autenticación o está en blanco
**Solución:** Configura `E2E_EMAIL` y `E2E_PASSWORD` en `.env.test`

### Warning: "Login failed - continuing as guest"
**Causa:** Credenciales incorrectas o endpoint de login cambió
**Solución:** Verifica credenciales en `.env.test` y que `/auth/login` exista

### Muchos "Clicks blocked"
**Causa:** Configuración muy restrictiva o UI usa clases `danger` en elementos seguros
**Solución:** Revisa `DANGEROUS_UI_CLASS_PATTERNS` y ajusta si es necesario

### Test se queda en loop infinito
**Causa:** `MAX_CLICKS_PER_PAGE` muy alto o página sin variedad de elementos
**Solución:** Reduce `MAX_CLICKS_PER_PAGE` a 5-10

---

## 📝 Notas de Desarrollo

### ¿Por Qué No Se Envían Formularios?

Para evitar crear datos innecesarios en la base de datos durante el test exploratorio. Si quieres probar envíos de formularios específicos, crea tests E2E dedicados (no monkey tests).

### ¿Puedo Agregar Más Acciones Peligrosas?

Sí, edita las constantes en `/frontend/tests/monkey.spec.ts`:
- `DANGEROUS_ENDPOINT_PATTERNS`
- `DANGEROUS_ROUTE_PATTERNS`
- `DANGEROUS_UI_TEXT_PATTERNS`
- `DANGEROUS_UI_CLASS_PATTERNS`

### ¿Funciona con NextAuth u otros Providers de Auth?

El test está diseñado para auth custom (Go backend). Para NextAuth:
1. Modifica la función `attemptLogin()`
2. Ajusta detección de login page si es diferente de `/auth/login`

---

## 🎯 Próximos Pasos

1. **Instalar dependencias:**
   ```bash
   npm install
   npx playwright install
   ```

2. **Crear archivo de variables de entorno:**
   ```bash
   cp .env.example .env.test
   # Editar .env.test con credenciales de prueba
   ```

3. **Ejecutar test con navegador visible:**
   ```bash
   npm run test:e2e:headed
   ```

4. **Revisar reporte:**
   - Consola: Reporte impreso al final del test
   - HTML: `npm run test:e2e:report`

---

## 📚 Referencias

- [Playwright Documentation](https://playwright.dev)
- [Playwright Test API](https://playwright.dev/docs/api/class-test)
- [IRIS Backend API Docs](../backend/README.md)

---

## ⚖️ Licencia y Responsabilidad

Este test está diseñado ÚNICAMENTE para entornos de desarrollo local.

**⚠️ NUNCA ejecutes este test contra:**
- Producción
- Staging con datos reales
- Entornos públicos/compartidos

El test incluye salvaguardas, pero la responsabilidad final de ejecutarlo en el entorno correcto es del desarrollador.

---

**¿Preguntas o Issues?** Abre un issue en el repositorio o contacta al equipo de QA.
