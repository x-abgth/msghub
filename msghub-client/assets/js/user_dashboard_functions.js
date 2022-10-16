// To change the chat head name according user interaction
function pmHim(uname) {
	const name = document.getElementById("chat-user-name");
	name.innerText = uname;
}

function openCreateGroup() {
	const chatPartHeader = document.getElementById("dashboard-chat-part-header");
	const chatPartBody = document.getElementById("dashboard-chat-part-body");
	const createGroupHeader = document.getElementById("dashboard-create-group-header");
	const createGroupBody = document.getElementById("dashboard-create-group-body");
	
	createGroupHeader.classList.remove("d-none");
	createGroupBody.classList.remove("d-none");

	chatPartHeader.classList.add("d-none");
	chatPartBody.classList.add("d-none");

	createGroupHeader.classList.add("d-block");
	createGroupBody.classList.add("d-block");
}

function closeCreateGroup() {
	const chatPartHeader = document.getElementById("dashboard-chat-part-header");
	const chatPartBody = document.getElementById("dashboard-chat-part-body");
	const createGroupHeader = document.getElementById("dashboard-create-group-header");
	const createGroupBody = document.getElementById("dashboard-create-group-body");
	
	chatPartHeader.classList.remove("d-none");
	chatPartBody.classList.remove("d-none");

	createGroupHeader.classList.add("d-none");
	createGroupBody.classList.add("d-none");

	chatPartHeader.classList.add("d-block");
	chatPartBody.classList.add("d-block");
}

// Upload image to the create group 

$("#profileImage").click(function(e) {
    $("#imageUpload").click();
});

function fasterPreview( uploader ) {
    if ( uploader.files && uploader.files[0] ){
          $('#profileImage').attr('src', 
             window.URL.createObjectURL(uploader.files[0]) );
    }
}

$("#imageUpload").change(function(){
    fasterPreview( this );
});

jQuery('#createNewGroupForm').validate({
	rules: {
	  groupName: {
	  	required: true,
	  },
	}, messages: {
	  groupName: 'Please enter a name for the group.',
	}, submitHandler: function (createNewGroupForm) {
	  createNewGroupForm.submit();
	}
});

// Ajax request to create group
