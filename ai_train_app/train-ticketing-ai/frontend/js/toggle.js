// Toggle between Traditional Booking and AI Assistant

document.addEventListener('DOMContentLoaded', () => {
    const traditionalPanel = document.getElementById('traditionalPanel');
    const aiPanel = document.getElementById('aiPanel');
    const switchToAI = document.getElementById('switchToAI');
    const switchToTraditional = document.getElementById('switchToTraditional');

    // Switch to AI Assistant
    switchToAI.addEventListener('click', () => {
        traditionalPanel.style.display = 'none';
        aiPanel.style.display = 'flex';
    });

    // Switch to Traditional Booking
    switchToTraditional.addEventListener('click', () => {
        aiPanel.style.display = 'none';
        traditionalPanel.style.display = 'flex';
    });
});
