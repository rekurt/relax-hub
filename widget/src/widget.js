(function() {
  'use strict';

  // Configuration
  const CONFIG = {
    apiBaseUrl: typeof window.BANI_WIDGET_API_URL !== 'undefined'
      ? window.BANI_WIDGET_API_URL
      : 'https://api.bani.ru',
    defaultLanguage: 'ru',
    defaultPrimaryColor: '#0f766e',
    defaultFontFamily: 'Manrope, "Avenir Next", "Segoe UI", ui-sans-serif, system-ui, sans-serif',
    translations: {
      ru: {
        title: 'Бронирование',
        selectDate: 'Выберите дату',
        selectTime: 'Выберите время',
        bookingForm: 'Данные для бронирования',
        name: 'Имя',
        phone: 'Телефон',
        email: 'Email',
        guests: 'Количество гостей',
        price: 'Цена',
        pricePerHour: 'за час',
        book: 'Забронировать',
        back: 'Назад',
        confirmation: 'Спасибо за бронирование!',
        bookingId: 'Номер бронирования:',
        dateTime: 'Дата и время:',
        totalPrice: 'Итого:',
        close: 'Закрыть',
        loading: 'Загрузка...',
        error: 'Ошибка',
        noSlotsAvailable: 'Нет доступных слотов на эту дату',
        invalidInput: 'Пожалуйста, заполните все поля',
      },
      en: {
        title: 'Booking',
        selectDate: 'Select date',
        selectTime: 'Select time',
        bookingForm: 'Booking details',
        name: 'Name',
        phone: 'Phone',
        email: 'Email',
        guests: 'Number of guests',
        price: 'Price',
        pricePerHour: 'per hour',
        book: 'Book',
        back: 'Back',
        confirmation: 'Thank you for booking!',
        bookingId: 'Booking ID:',
        dateTime: 'Date and time:',
        totalPrice: 'Total:',
        close: 'Close',
        loading: 'Loading...',
        error: 'Error',
        noSlotsAvailable: 'No available slots for this date',
        invalidInput: 'Please fill in all fields',
      }
    }
  };

  function sanitizeColor(value) {
    return /^#[0-9a-fA-F]{3,8}$/.test(value) ? value : null;
  }

  function sanitizeFontFamily(value) {
    return value.replace(/[^a-zA-Z0-9\s,\-]/g, '');
  }

  function escapeHTMLAttr(str) {
    return str.replace(/&/g, '&amp;').replace(/'/g, '&#39;').replace(/"/g, '&quot;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }

  class BaniWidget {
    constructor(element) {
      this.element = element;
      this.apiKey = element.getAttribute('data-api-key');
      this.primaryColor = sanitizeColor(element.getAttribute('data-color')) || CONFIG.defaultPrimaryColor;
      this.fontFamily = sanitizeFontFamily(element.getAttribute('data-font-family') || CONFIG.defaultFontFamily);
      this.language = element.getAttribute('data-language') || CONFIG.defaultLanguage;
      this.t = CONFIG.translations[this.language] || CONFIG.translations[CONFIG.defaultLanguage];

      if (!this.apiKey) {
        console.error('Bani Widget: data-api-key attribute is required');
        return;
      }

      this.state = {
        view: 'date-selection',
        selectedDate: null,
        selectedSlot: null,
        bathhouse: null,
        slots: [],
        loading: false,
        error: null,
      };

      this.init();
    }

    init() {
      this.setupStyles();
      this.render();
      this.fetchBathhouse();
    }

    setupStyles() {
      this.element.style.setProperty('--bani-widget-primary', this.primaryColor);
      this.element.style.setProperty('--bani-widget-primary-strong', this.adjustColor(this.primaryColor, -18));
      this.element.style.setProperty('--bani-widget-font-family', this.fontFamily);
    }

    adjustColor(color, percent) {
      const num = parseInt(color.replace('#',''), 16);
      const amt = Math.round(2.55 * percent);
      const R = (num >> 16) + amt;
      const G = (num >> 8 & 0x00FF) + amt;
      const B = (num & 0x0000FF) + amt;
      return '#' + (0x1000000 + (R<255?R<1?0:R:255)*0x10000 +
        (G<255?G<1?0:G:255)*0x100 + (B<255?B<1?0:B:255))
        .toString(16).slice(1);
    }

    async fetchBathhouse() {
      try {
        this.state.loading = true;
        this.render();

        const response = await fetch(`${CONFIG.apiBaseUrl}/api/v1/widget/${this.apiKey}/bathhouse`);
        if (!response.ok) {
          throw new Error('Failed to fetch bathhouse info');
        }

        const res = await response.json();
        this.state.bathhouse = res.data;
        this.state.error = null;
      } catch (error) {
        this.state.error = error.message;
        console.error('Bani Widget Error:', error);
      } finally {
        this.state.loading = false;
        this.render();
      }
    }

    async fetchSlots(date) {
      try {
        this.state.loading = true;
        this.render();

        const dateStr = date.toISOString().split('T')[0];
        const response = await fetch(
          `${CONFIG.apiBaseUrl}/api/v1/widget/${this.apiKey}/slots?date=${dateStr}`
        );

        if (!response.ok) {
          throw new Error('Failed to fetch available slots');
        }

        const res = await response.json();
        this.state.slots = res.data || [];
        this.state.selectedDate = date;
        this.state.view = 'slot-selection';
        this.state.error = null;
      } catch (error) {
        this.state.error = error.message;
        console.error('Bani Widget Error:', error);
      } finally {
        this.state.loading = false;
        this.render();
      }
    }

    async submitBooking(formData) {
      try {
        this.state.loading = true;
        this.render();

        const slot = this.state.selectedSlot;
        const payload = {
          name: formData.name,
          phone: formData.phone,
          email: formData.email,
          start_time: slot.start_time,
          end_time: slot.end_time,
          guest_count: parseInt(formData.guests),
          comment: formData.comment || ''
        };

        const response = await fetch(
          `${CONFIG.apiBaseUrl}/api/v1/widget/${this.apiKey}/booking`,
          {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
          }
        );

        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.error?.message || 'Failed to create booking');
        }

        const res = await response.json();
        this.state.selectedSlot = { ...slot, ...res.data };
        this.state.view = 'confirmation';
        this.state.error = null;
        this.state.loading = false;
        this.render();
      } catch (error) {
        this.state.error = error.message;
        console.error('Bani Widget Error:', error);
        this.state.loading = false;
        this.render();
      }
    }

    formatPrice(price) {
      return (price / 100).toLocaleString(this.language, {
        minimumFractionDigits: 0,
        maximumFractionDigits: 2
      });
    }

    render() {
      const content = this.renderCurrentView();
      this.element.innerHTML = content;
      this.attachEventListeners();
    }

    renderCurrentView() {
      if (this.state.error) {
        return `<div class="bani-widget"><div class="bani-widget-error">${this.t.error}: ${this.escapeHtml(this.state.error)}</div></div>`;
      }

      if (this.state.loading && !this.state.bathhouse) {
        return `<div class="bani-widget"><div class="bani-widget-loading">${this.t.loading}</div></div>`;
      }

      switch (this.state.view) {
        case 'date-selection':
          return this.renderDateSelection();
        case 'slot-selection':
          return this.renderSlotSelection();
        case 'booking-form':
          return this.renderBookingForm();
        case 'confirmation':
          return this.renderConfirmation();
        default:
          return '';
      }
    }

    escapeHtml(text) {
      const map = {'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#039;'};
      return text.replace(/[&<>"']/g, m => map[m]);
    }

    renderDateSelection() {
      if (!this.state.bathhouse) return '';

      const today = new Date();
      const daysToShow = 30;

      let html = `<div class="bani-widget">`;
      html += `<h2>${this.t.title}</h2>`;
      html += this.renderBathhouseInfo();

      html += `<div class="bani-widget-calendar">`;
      html += `<h3>${this.t.selectDate}</h3>`;
      html += this.renderCalendar(today, daysToShow);
      html += `</div>`;
      html += `</div>`;

      return html;
    }

    renderBathhouseInfo() {
      if (!this.state.bathhouse) return '';
      const bathhouse = this.state.bathhouse;
      return `
        <div class="bani-widget-bathhouse-info">
          <div class="bani-widget-bathhouse-name">${this.escapeHtml(bathhouse.name)}</div>
          <div class="bani-widget-bathhouse-address">${this.escapeHtml(bathhouse.address)}</div>
          <div class="bani-widget-bathhouse-price">
            ${this.formatPrice(bathhouse.price_per_hour)} ${this.t.pricePerHour}
          </div>
        </div>
      `;
    }

    renderCalendar(startDate, daysToShow) {
      const days = ['ПН', 'ВТ', 'СР', 'ЧТ', 'ПТ', 'СБ', 'ВС'];

      let html = `<div class="bani-widget-calendar-header">`;
      for (let i = 0; i < 7; i++) {
        const dayLabel = this.language === 'ru' ? days[i] : ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'][i];
        html += `<div class="bani-widget-calendar-day">${dayLabel}</div>`;
      }
      html += `</div>`;

      html += `<div class="bani-widget-calendar-grid">`;
      for (let i = 0; i < daysToShow; i++) {
        const date = new Date(startDate);
        date.setDate(date.getDate() + i);
        const dateStr = date.toISOString().split('T')[0];

        const isSelected = this.state.selectedDate &&
          this.state.selectedDate.toISOString().split('T')[0] === dateStr;

        html += `<button
          class="bani-widget-calendar-date ${isSelected ? 'selected' : ''}"
          data-date="${dateStr}"
        >
          ${date.getDate()}
        </button>`;
      }
      html += `</div>`;

      return html;
    }

    renderSlotSelection() {
      if (!this.state.bathhouse) return '';

      let html = `<div class="bani-widget">`;
      html += `<h2>${this.t.title}</h2>`;
      html += this.renderBathhouseInfo();
      html += `
        <div class="bani-widget-nav">
          <button class="bani-widget-button bani-widget-button-secondary" data-action="back">
            ${this.t.back}
          </button>
        </div>
      `;

      html += `<div class="bani-widget-selection-summary">`;
      html += `<strong>${this.t.selectTime}</strong>`;
      if (this.state.selectedDate) {
        const dateStr = this.state.selectedDate.toLocaleDateString(this.language);
        html += `<div class="bani-widget-muted">${dateStr}</div>`;
      }
      html += `</div>`;

      if (this.state.slots.length === 0) {
        html += `<div class="bani-widget-empty">`;
        html += this.t.noSlotsAvailable;
        html += `</div>`;
      } else {
        html += `<div class="bani-widget-slots">`;
        this.state.slots.forEach(slot => {
          const isSelected = this.state.selectedSlot &&
            this.state.selectedSlot.start_time === slot.start_time;
          const isAvailable = slot.available;

          const startTime = new Date(slot.start_time);
          const endTime = new Date(slot.end_time);
          const timeStr = `${startTime.getHours().toString().padStart(2, '0')}:${startTime.getMinutes().toString().padStart(2, '0')} - ${endTime.getHours().toString().padStart(2, '0')}:${endTime.getMinutes().toString().padStart(2, '0')}`;

          html += `
            <button
              class="bani-widget-slot ${isSelected ? 'selected' : ''} ${!isAvailable ? 'disabled' : ''}"
              data-slot='${escapeHTMLAttr(JSON.stringify(slot))}'
              ${!isAvailable ? 'disabled' : ''}
            >
              <span class="bani-widget-slot-time">${timeStr}</span>
              <span class="bani-widget-slot-price">${this.formatPrice(slot.price)}</span>
            </button>
          `;
        });
        html += `</div>`;

        if (this.state.selectedSlot) {
          html += `
            <button class="bani-widget-button" data-action="to-form">
              ${this.t.bookingForm}
            </button>
          `;
        }
      }

      html += `</div>`;
      return html;
    }

    renderBookingForm() {
      if (!this.state.bathhouse || !this.state.selectedSlot) return '';

      const slot = this.state.selectedSlot;
      const startTime = new Date(slot.start_time);
      const endTime = new Date(slot.end_time);

      const timeStr = `${startTime.getHours().toString().padStart(2, '0')}:${startTime.getMinutes().toString().padStart(2, '0')} - ${endTime.getHours().toString().padStart(2, '0')}:${endTime.getMinutes().toString().padStart(2, '0')}`;
      const dateStr = startTime.toLocaleDateString(this.language);

      let html = `<div class="bani-widget">`;
      html += `<h2>${this.t.bookingForm}</h2>`;
      html += this.renderBathhouseInfo();

      html += `
        <div class="bani-widget-selection-summary">
          <div class="bani-widget-muted">${this.t.selectTime}</div>
          <div style="font-size: 16px; font-weight: 800; color: var(--bani-widget-text);">${dateStr}</div>
          <div style="font-size: 16px; font-weight: 800; color: var(--bani-widget-text);">${timeStr}</div>
          <div class="bani-widget-muted" style="margin-top: 8px;">
            ${this.t.totalPrice}: <strong>${this.formatPrice(slot.price)}</strong>
          </div>
        </div>
      `;

      html += `
        <form id="bani-booking-form" style="margin-bottom: 16px;">
          <div class="bani-widget-form-group">
            <label class="bani-widget-label">${this.t.name}</label>
            <input type="text" class="bani-widget-input" name="name" required />
          </div>
          <div class="bani-widget-form-group">
            <label class="bani-widget-label">${this.t.phone}</label>
            <input type="tel" class="bani-widget-input" name="phone" required />
          </div>
          <div class="bani-widget-form-group">
            <label class="bani-widget-label">${this.t.email}</label>
            <input type="email" class="bani-widget-input" name="email" required />
          </div>
          <div class="bani-widget-form-group">
            <label class="bani-widget-label">${this.t.guests}</label>
            <input type="number" class="bani-widget-input" name="guests" min="1" max="${this.state.bathhouse.max_guests}" required />
          </div>
          <div class="bani-widget-nav">
            <button type="button" class="bani-widget-button bani-widget-button-secondary" data-action="back">
              ${this.t.back}
            </button>
            <button type="submit" class="bani-widget-button" ${this.state.loading ? 'disabled' : ''}>
              ${this.state.loading ? this.t.loading : this.t.book}
            </button>
          </div>
        </form>
        </div>
      `;

      return html;
    }

    renderConfirmation() {
      const slot = this.state.selectedSlot;
      const bathhouse = this.state.bathhouse;
      const startTime = new Date(slot.start_time);
      const endTime = new Date(slot.end_time);

      const timeStr = startTime.toLocaleDateString(this.language) + ' ' +
        startTime.toLocaleTimeString(this.language, { hour: '2-digit', minute: '2-digit' }) + ' - ' +
        endTime.toLocaleTimeString(this.language, { hour: '2-digit', minute: '2-digit' });

      let html = `<div class="bani-widget"><div class="bani-widget-confirmation">`;
      html += `<div class="bani-widget-confirmation-icon">✓</div>`;
      html += `<div class="bani-widget-confirmation-message">${this.t.confirmation}</div>`;
      html += `
        <div class="bani-widget-confirmation-details">
          <div class="bani-widget-confirmation-detail">
            <span>${this.t.bookingId}</span>
            <strong>${this.escapeHtml(String(slot.id))}</strong>
          </div>
          <div class="bani-widget-confirmation-detail">
            <span>${this.escapeHtml(bathhouse.name)}</span>
          </div>
          <div class="bani-widget-confirmation-detail">
            <span>${this.t.dateTime}</span>
            <span>${timeStr}</span>
          </div>
          <div class="bani-widget-confirmation-detail">
            <span>${this.t.totalPrice}</span>
            <strong>${this.formatPrice(slot.price)}</strong>
          </div>
        </div>
      `;
      html += `<button class="bani-widget-button" data-action="close">${this.t.close}</button>`;
      html += `</div></div>`;

      return html;
    }

    attachEventListeners() {
      const widget = this.element.querySelector('.bani-widget');

      widget?.querySelectorAll('[data-date]').forEach(btn => {
        btn.addEventListener('click', (e) => {
          e.preventDefault();
          const dateStr = btn.getAttribute('data-date');
          const date = new Date(dateStr + 'T00:00:00');
          this.fetchSlots(date);
        });
      });

      widget?.querySelectorAll('[data-slot]').forEach(btn => {
        btn.addEventListener('click', (e) => {
          e.preventDefault();
          const slotData = JSON.parse(btn.getAttribute('data-slot'));
          this.state.selectedSlot = slotData;
          this.render();
        });
      });

      const form = widget?.querySelector('#bani-booking-form');
      if (form) {
        form.addEventListener('submit', async (e) => {
          e.preventDefault();
          const formData = new FormData(form);

          if (!formData.get('name') || !formData.get('phone') || !formData.get('email')) {
            this.state.error = this.t.invalidInput;
            this.render();
            return;
          }

          await this.submitBooking({
            name: formData.get('name'),
            phone: formData.get('phone'),
            email: formData.get('email'),
            guests: formData.get('guests'),
            comment: formData.get('comment') || ''
          });
        });
      }

      widget?.querySelectorAll('[data-action]').forEach(btn => {
        btn.addEventListener('click', (e) => {
          e.preventDefault();
          const action = btn.getAttribute('data-action');

          switch (action) {
            case 'back':
              if (this.state.view === 'slot-selection') {
                this.state.selectedDate = null;
                this.state.selectedSlot = null;
                this.state.slots = [];
                this.state.view = 'date-selection';
              } else if (this.state.view === 'booking-form') {
                this.state.selectedSlot = null;
                this.state.view = 'slot-selection';
              }
              this.render();
              break;
            case 'to-form':
              this.state.view = 'booking-form';
              this.render();
              break;
            case 'close':
              this.state.view = 'date-selection';
              this.state.selectedDate = null;
              this.state.selectedSlot = null;
              this.state.slots = [];
              this.render();
              break;
          }
        });
      });
    }
  }

  function initializeWidgets() {
    document.querySelectorAll('[id="bani-widget"]').forEach(element => {
      if (!element.dataset.initialized) {
        new BaniWidget(element);
        element.dataset.initialized = 'true';
      }
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeWidgets);
  } else {
    initializeWidgets();
  }

  window.BaniWidget = BaniWidget;
})();
