// Internationalization (i18n) - English and Italian translations

const translations = {
    en: {
        subtitle: "Book Your Train Tickets",
        traditional_booking: "🎫 Traditional Booking",
        ai_assistant: "🤖 AI Assistant",
        from: "From:",
        to: "To:",
        date: "Date:",
        select_departure: "Select departure station",
        select_arrival: "Select arrival station",
        time_preference: "Time Preference:",
        any_time: "Any Time",
        morning: "Morning (6-12)",
        afternoon: "Afternoon (12-18)",
        evening: "Evening (18-24)",
        adults: "Adults:",
        seniors: "Seniors (65+):",
        children: "Children (4-17):",
        wifi_required: "WiFi Required",
        food_required: "Food Service Required",
        search_trains: "Search Trains",
        ai_greeting: "Hello! I'm your AI booking assistant. I can help you find and book train tickets.",
        ai_try_saying: "Try saying something like:",
        ai_example_1: '"I need to go from Milan to Rome tomorrow morning"',
        ai_example_2: '"Show me trains from Venice to Florence next week"',
        ai_example_3: '"Book 2 tickets on the 7:30 train"',
        chat_placeholder: "Type your message here...",
        send: "Send",
        loading: "Loading...",
        cancel: "Cancel",
        confirm_booking: "Confirm Booking",
        booking_confirmed: "Booking Confirmed!",
        close: "Close",
        complete_booking: "Complete Your Booking"
    },
    it: {
        subtitle: "Prenota i Tuoi Biglietti del Treno",
        traditional_booking: "🎫 Prenotazione Tradizionale",
        ai_assistant: "🤖 Assistente AI",
        from: "Da:",
        to: "A:",
        date: "Data:",
        select_departure: "Seleziona stazione di partenza",
        select_arrival: "Seleziona stazione di arrivo",
        time_preference: "Preferenza Oraria:",
        any_time: "Qualsiasi Orario",
        morning: "Mattina (6-12)",
        afternoon: "Pomeriggio (12-18)",
        evening: "Sera (18-24)",
        adults: "Adulti:",
        seniors: "Anziani (65+):",
        children: "Bambini (4-17):",
        wifi_required: "WiFi Richiesto",
        food_required: "Servizio Ristorazione Richiesto",
        search_trains: "Cerca Treni",
        ai_greeting: "Ciao! Sono il tuo assistente AI per le prenotazioni. Posso aiutarti a trovare e prenotare biglietti del treno.",
        ai_try_saying: "Prova a dire qualcosa come:",
        ai_example_1: '"Ho bisogno di andare da Milano a Roma domani mattina"',
        ai_example_2: '"Mostrami i treni da Venezia a Firenze la prossima settimana"',
        ai_example_3: '"Prenota 2 biglietti sul treno delle 7:30"',
        chat_placeholder: "Scrivi il tuo messaggio qui...",
        send: "Invia",
        loading: "Caricamento...",
        cancel: "Annulla",
        confirm_booking: "Conferma Prenotazione",
        booking_confirmed: "Prenotazione Confermata!",
        close: "Chiudi",
        complete_booking: "Completa la Tua Prenotazione"
    }
};

let currentLanguage = 'en';

// Initialize i18n on page load
document.addEventListener('DOMContentLoaded', () => {
    const languageSelect = document.getElementById('languageSelect');

    // Load saved language preference
    const savedLang = localStorage.getItem('language') || 'en';
    currentLanguage = savedLang;
    languageSelect.value = savedLang;

    // Apply initial translation
    translatePage(savedLang);

    // Listen for language changes
    languageSelect.addEventListener('change', (e) => {
        const newLang = e.target.value;
        currentLanguage = newLang;
        localStorage.setItem('language', newLang);
        translatePage(newLang);
    });
});

// Translate all elements with data-i18n attribute
function translatePage(lang) {
    const elements = document.querySelectorAll('[data-i18n]');

    elements.forEach(element => {
        const key = element.getAttribute('data-i18n');
        if (translations[lang] && translations[lang][key]) {
            element.textContent = translations[lang][key];
        }
    });

    // Handle placeholders separately
    const placeholderElements = document.querySelectorAll('[data-i18n-placeholder]');
    placeholderElements.forEach(element => {
        const key = element.getAttribute('data-i18n-placeholder');
        if (translations[lang] && translations[lang][key]) {
            element.placeholder = translations[lang][key];
        }
    });
}

// Export for use in other scripts
window.i18n = {
    translate: (key) => {
        return translations[currentLanguage][key] || key;
    },
    getCurrentLanguage: () => currentLanguage
};
