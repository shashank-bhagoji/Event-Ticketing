const API = "https://event-ticketing-4v0x.onrender.com";
let currentUser = null;

async function handleLogin(isOrganizer) {
  const email = document.getElementById("loginEmail").value;
  const password = document.getElementById("loginPassword").value;

  if (!email || !password) {
    alert("Please enter both email and password");
    return;
  }

  try {
    const res = await fetch(`${API}/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        email: email,
        password: password,
        is_organizer: isOrganizer
      })
    });

    const data = await res.json();
    if (!res.ok) {
      alert(data.error || "Login failed");
      return;
    }

    currentUser = data.user;

    // Switch Views
    document.getElementById("login-section").classList.add("hidden");
    document.getElementById("main-content").classList.remove("hidden");

    if (currentUser.Role === "organizer") {
      document.getElementById("create-event-section").classList.remove("hidden");
    } else {
      document.getElementById("create-event-section").classList.add("hidden");
    }

    loadEvents();
  } catch (err) {
    console.error("Login error", err);
    alert("An error occurred during login. Check console.");
  }
}

async function handleCreateEvent() {
  const title = document.getElementById("title").value;
  const capacity = document.getElementById("capacity").value;
  const eventDate = document.getElementById("eventDate").value;
  const eventTime = document.getElementById("eventTime").value;

  if (!title || !capacity || !eventDate || !eventTime) {
    alert("Please fill in all fields.");
    return;
  }

  // Combine date and time
  const combinedDateTime = new Date(`${eventDate}T${eventTime}`);

  if (combinedDateTime <= new Date()) {
    alert("Event time must be in the future.");
    return;
  }

  await fetch(`${API}/events`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      title: title,
      description: "Sample Event",
      eventdate: combinedDateTime.toISOString(), // Send to Go backend mapping
      totalCapacity: parseInt(capacity),
      createdBy: currentUser.ID
    })
  });

  document.getElementById("title").value = "";
  document.getElementById("capacity").value = "";
  document.getElementById("eventDate").value = "";
  document.getElementById("eventTime").value = "";

  loadEvents();
}

async function loadEvents() {
  const res = await fetch(`${API}/events`);
  let events = [];
  try {
    events = await res.json();
  } catch (e) {
    console.error("Failed to parse events JSON", e);
  }

  const container = document.getElementById("events");
  container.innerHTML = "";

  if (!events || events.length === 0) {
    container.innerHTML = `<div class="no-events">No events available right now. Be the first to create one!</div>`;
    return;
  }

  events.forEach(event => {
    const div = document.createElement("div");
    div.className = "event-card";

    const formattedDate = new Date(event.EventDate).toLocaleString();

    div.innerHTML = `
      <div class="event-details">
        <h3>${event.Title}</h3>
        <p style="color: var(--text-muted); font-size: 0.9rem; margin-top: -0.5rem; display: flex; align-items: center; gap: 0.5rem;">
          <svg style="min-width: 14px" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
          ${formattedDate}
        </p>
        <p class="capacity-badge">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M22 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
          Available: <span>${event.AvailableCapacity}</span>
        </p>
      </div>
      <div style="display: flex; gap: 1rem;">
        ${currentUser && currentUser.Role === 'user' ? (event.AvailableCapacity > 0 ? `<button class="btn btn-secondary" onclick="openRegisterModal(${event.ID})">Register Now</button>` : `<button class="btn btn-secondary" style="background:#e5e7eb; color:#9ca3af; border-color:#e5e7eb; cursor:not-allowed;" disabled>No available slots</button>`) : ''}
        ${currentUser && currentUser.Role === 'organizer' ? `<button class="btn btn-primary" style="background: white; border: 1px solid #ef4444; color: #ef4444;" onclick="deleteEvent(${event.ID})">Delete Event</button>` : ''}
      </div>
    `;
    container.appendChild(div);
  });
}

let currentEventId = null;

function openRegisterModal(eventId) {
  currentEventId = eventId;
  const modal = document.getElementById('register-modal');
  modal.classList.remove('hidden');
}

function closeModal() {
  currentEventId = null;
  const modal = document.getElementById('register-modal');
  modal.classList.add('hidden');
}

async function submitRegistration() {
  if (!currentUser) return;
  const userId = currentUser.ID;

  const res = await fetch(`${API}/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      user_id: userId,
      event_id: currentEventId
    })
  });

  const data = await res.json();
  if (!res.ok) {
    alert(data.error || "Registration failed.");
  } else {
    alert(data.message || "Registration successful!");
  }

  closeModal();
  loadEvents();
}

// Replaced loadEvents call, since we only load events after login.

async function deleteEvent(eventId) {
  if (!confirm("Are you sure you want to delete this event? This action cannot be undone.")) return;

  await fetch(`${API}/events/${eventId}`, {
    method: "DELETE"
  });

  loadEvents();
}