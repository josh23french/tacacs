package main

import (
	"context"
	"fmt"
	"log"

	"github.com/josh23french/tacacs/pkg/args"
	"github.com/josh23french/tacacs/pkg/authorizer"
	"github.com/josh23french/tacacs/pkg/config"
	"github.com/josh23french/tacacs/pkg/backends"
	"github.com/josh23french/tacplus"
)

type Handler struct {
	Authenticator *Authenticator
	Authorizer *authorizer.Authorizer
}

func NewHandler(cfg *config.Config, backends *backends.BackendManager) *Handler {
	return &Handler{
		Authenticator: NewAuthenticator(cfg),
		Authorizer: authorizer.New(cfg, backends),
	}
}

func (h *Handler) HandleAcctRequest(c context.Context, req *tacplus.AcctRequest) *tacplus.AcctReply {
	log.Printf("HandleAcctRequest: %v\n", req)
	return nil
}

func (h *Handler) HandleAuthenStart(c context.Context, req *tacplus.AuthenStart, sess *tacplus.ServerSession) *tacplus.AuthenReply {
	log.Printf("HandleAuthenStart: %v\n", req)
	cont, err := sess.GetPass(c, "Password:")
	if err != nil {
		fmt.Printf("Error GetPass: %v\n", err)
		return nil
	}
	if cont.Abort {
		fmt.Printf("GetPass aborted: %v\n", cont.Message)
		return nil
	}
	pass := cont.Message

	auth_ok, err := h.Authenticator.Auth(req.User, pass)

	if err != nil {
		fmt.Printf("Error Authenticating %v: %v\n", req.User, fmt.Errorf("%w", err))
		return &tacplus.AuthenReply{
			Status:    tacplus.AuthenStatusError,
			ServerMsg: fmt.Sprintf("%v", fmt.Errorf("Error: %w", err)),
		}
	}

	if auth_ok {
		return &tacplus.AuthenReply{
			Status: tacplus.AuthenStatusPass,
		}
	}
	return &tacplus.AuthenReply{
		Status: tacplus.AuthenStatusFail,
	}
}

func (h *Handler) HandleAuthorRequest(c context.Context, req *tacplus.AuthorRequest) *tacplus.AuthorResponse {
	log.Printf("HandleAuthorRequest: %v\n", req)
	log.Printf("Priv-lvl: %v", req.PrivLvl)
	mp := args.ParseAuthorArgs(req.Arg)
	if mp != nil {
		if mp.Service != nil && mp.Cmd != nil {
		    if *mp.Service == "shell" {
		        auth_ok, err := h.Authorizer.Auth(req.User, mp)
		        if err != nil {
		            return &tacplus.AuthorResponse{Status: tacplus.AuthorStatusFail}
		        }
		        if auth_ok {
		            log.Printf("Allowing %v to run %v", req.User, mp.AsShellCommand())
		            arg := mp.Marshall()
		            log.Printf("ARGS: %+v\n", arg)
			        return &tacplus.AuthorResponse{
			            Status: tacplus.AuthorStatusPassAdd,
			         //   Arg: arg,
			        }
		        }
	            log.Printf("Explicitly denying %v to run %v", req.User, mp.AsShellCommand())
		        return &tacplus.AuthorResponse{Status: tacplus.AuthorStatusFail}
		    }
		}

		return &tacplus.AuthorResponse{Status: tacplus.AuthorStatusFail}
	}
	return &tacplus.AuthorResponse{Status: tacplus.AuthorStatusFail}
}
